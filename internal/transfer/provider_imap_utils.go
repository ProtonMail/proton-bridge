// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package transfer

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	imapID "github.com/ProtonMail/go-imap-id"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"
	"github.com/emersion/go-sasl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	imapDialTimeout      = 5 * time.Second
	imapRetries          = 10
	imapReconnectTimeout = 30 * time.Minute
	imapReconnectSleep   = time.Minute
)

type imapErrorLogger struct {
	log *logrus.Entry
}

func (l *imapErrorLogger) Printf(f string, v ...interface{}) {
	l.log.Errorf(f, v...)
}

func (l *imapErrorLogger) Println(v ...interface{}) {
	l.log.Errorln(v...)
}

func imapClientDial(addr string) (IMAPClientProvider, error) {
	if _, err := net.DialTimeout("tcp", addr, imapDialTimeout); err != nil {
		return nil, errors.Wrap(err, "failed to dial server")
	}

	client, err := imapClientDialHelper(addr)
	if err == nil {
		client.ErrorLog = &imapErrorLogger{logrus.WithField("pkg", "imap-client")}
		// Logrus `WriterLevel` fails for big messages because of bufio.MaxScanTokenSize limit.
		// Also, this spams a lot, uncomment once needed during development.
		//client.SetDebug(imap.NewDebugWriter(
		//	logrus.WithField("pkg", "imap/client").WriterLevel(logrus.TraceLevel),
		//	logrus.WithField("pkg", "imap/server").WriterLevel(logrus.TraceLevel),
		//))
	}
	return client, err
}

func imapClientDialHelper(addr string) (*imapClient.Client, error) {
	host, _, _ := net.SplitHostPort(addr)
	if host == "127.0.0.1" {
		return imapClient.Dial(addr)
	}

	// IMAP mail.yahoo.com has problem with golang TLS 1.3 implementation
	// with weird behaviour, i.e., Yahoo does not return error during dial
	// or handshake but server does logs out right after successful login
	// leaving no time to perform any action.
	// Limiting TLS to version 1.2 is working just fine.
	var tlsConf *tls.Config
	if strings.Contains(strings.ToLower(host), "yahoo") {
		log.Warning("Yahoo server detected: limiting maximal TLS version to 1.2.")
		tlsConf = &tls.Config{MaxVersion: tls.VersionTLS12}
	}
	return imapClient.DialTLS(addr, tlsConf)
}

func (p *IMAPProvider) ensureConnection(callback func() error) error {
	return p.ensureConnectionAndSelection(callback, "")
}

func (p *IMAPProvider) ensureConnectionAndSelection(callback func() error, ensureSelectedIn string) error {
	var callErr error
	for i := 1; i <= imapRetries; i++ {
		callErr = callback()
		if callErr == nil {
			return nil
		}

		log.WithField("attempt", i).WithError(callErr).Warning("IMAP call failed, trying reconnect")
		err := p.tryReconnect(ensureSelectedIn)
		if err != nil {
			return err
		}
	}
	return errors.Wrap(callErr, "too many retries")
}

func (p *IMAPProvider) tryReconnect(ensureSelectedIn string) error {
	start := time.Now()
	var previousErr error
	for {
		if time.Since(start) > imapReconnectTimeout {
			return previousErr
		}

		err := pmapi.CheckConnection()
		log.WithError(err).Debug("Connection check")
		if err != nil {
			time.Sleep(imapReconnectSleep)
			previousErr = err
			continue
		}

		err = p.reauth()
		log.WithError(err).Debug("Reauth")
		if err != nil {
			time.Sleep(imapReconnectSleep)
			previousErr = err
			continue
		}

		if ensureSelectedIn != "" {
			_, err = p.client.Select(ensureSelectedIn, true)
			log.WithError(err).Debug("Reselect")
			if err != nil {
				previousErr = err
				continue
			}
		}

		break
	}
	return nil
}

func (p *IMAPProvider) reauth() error {
	var state imap.ConnState

	// In some cases once go-imap fails, we cannot issue another command
	// because it would dead-lock. Let's simply ignore it, we want to open
	// new connection anyway.
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		if _, err := p.client.Capability(); err != nil {
			state = p.client.State()
		}
	}()
	select {
	case <-ch:
	case <-time.After(30 * time.Second):
	}

	log.WithField("addr", p.addr).WithField("state", state).Debug("Reconnecting")
	p.client = nil
	return p.auth()
}

func (p *IMAPProvider) auth() error { //nolint[funlen]
	log := log.WithField("addr", p.addr)

	log.Info("Connecting to server")

	client, err := p.clientDialer(p.addr)
	if err != nil {
		return ErrIMAPConnection{imapError{Err: err, Message: "failed to connect to server"}}
	}
	p.client = client

	log.Info("Connected")

	if (p.client.State() & imap.AuthenticatedState) == imap.AuthenticatedState {
		return nil
	}

	capability, err := p.client.Capability()
	log.WithField("capability", capability).WithError(err).Debug("Server capability")
	if err != nil {
		return ErrIMAPConnection{imapError{Err: err, Message: "failed to get capabilities"}}
	}

	// SASL AUTH PLAIN
	if ok, _ := p.client.SupportAuth("PLAIN"); p.client.State() == imap.NotAuthenticatedState && ok {
		log.Debug("Trying plain auth")
		authPlain := sasl.NewPlainClient("", p.username, p.password)
		if err = p.client.Authenticate(authPlain); err != nil {
			return ErrIMAPAuth{imapError{Err: err, Message: "plain auth failed"}}
		}
	}

	// LOGIN: if the server reports the IMAP4rev1 capability then it is standards conformant and must support login.
	if ok, _ := p.client.Support("IMAP4rev1"); p.client.State() == imap.NotAuthenticatedState && ok {
		log.Debug("Trying login")
		if err = p.client.Login(p.username, p.password); err != nil {
			return ErrIMAPAuth{imapError{Err: err, Message: "login failed"}}
		}
	}

	if p.client.State() == imap.NotAuthenticatedState {
		return ErrIMAPAuthMethod{imapError{Err: err, Message: "unknown auth method"}}
	}

	log.Info("Logged in")

	if c, ok := p.client.(*imapClient.Client); ok {
		idClient := imapID.NewClient(c)
		if ok, err := idClient.SupportID(); err == nil && ok {
			serverID, err := idClient.ID(imapID.ID{
				imapID.FieldName:    "ImportExport",
				imapID.FieldVersion: constants.Version,
			})
			log.WithField("ID", serverID).WithError(err).Debug("Server info")
		}
	}

	return err
}

func (p *IMAPProvider) list() (mailboxes []*imap.MailboxInfo, err error) {
	err = p.ensureConnection(func() error {
		mailboxesCh := make(chan *imap.MailboxInfo)
		doneCh := make(chan error)

		go func() {
			doneCh <- p.client.List("", "*", mailboxesCh)
		}()

		for mailbox := range mailboxesCh {
			mailboxes = append(mailboxes, mailbox)
		}

		return <-doneCh
	})
	return
}

func (p *IMAPProvider) selectIn(mailboxName string) (mailbox *imap.MailboxStatus, err error) {
	err = p.ensureConnection(func() error {
		mailbox, err = p.client.Select(mailboxName, true)
		return err
	})
	return
}

func (p *IMAPProvider) fetch(ensureSelectedIn string, seqSet *imap.SeqSet, items []imap.FetchItem, processMessageCallback func(m *imap.Message)) error {
	return p.fetchHelper(false, ensureSelectedIn, seqSet, items, processMessageCallback)
}

func (p *IMAPProvider) uidFetch(ensureSelectedIn string, seqSet *imap.SeqSet, items []imap.FetchItem, processMessageCallback func(m *imap.Message)) error {
	return p.fetchHelper(true, ensureSelectedIn, seqSet, items, processMessageCallback)
}

func (p *IMAPProvider) fetchHelper(uid bool, ensureSelectedIn string, seqSet *imap.SeqSet, items []imap.FetchItem, processMessageCallback func(m *imap.Message)) error {
	return p.ensureConnectionAndSelection(func() error {
		messagesCh := make(chan *imap.Message)
		doneCh := make(chan error)

		go func() {
			if uid {
				doneCh <- p.client.UidFetch(seqSet, items, messagesCh)
			} else {
				doneCh <- p.client.Fetch(seqSet, items, messagesCh)
			}
		}()

		for message := range messagesCh {
			processMessageCallback(message)
		}

		err := <-doneCh
		return err
	}, ensureSelectedIn)
}
