// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package mocks

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-rfc5322"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SMTPClient struct {
	lock     *sync.Mutex
	debug    *debug
	t        TestingT
	conn     net.Conn
	response *bufio.Reader
	address  string
}

func NewSMTPClient(t TestingT, tag, smtpAddr string) *SMTPClient {
	conn, err := net.Dial("tcp", smtpAddr)
	require.NoError(t, err)
	if err != nil {
		return &SMTPClient{}
	}
	logrus.WithField("addr", conn.LocalAddr().String()).
		WithField("tag", tag).
		Debug("SMTP Dialed")
	response := bufio.NewReader(conn)

	// Read first response to opening connection.
	_, err = response.ReadString('\n')
	assert.NoError(t, err)

	return &SMTPClient{
		lock:     &sync.Mutex{},
		debug:    newDebug(tag),
		t:        t,
		conn:     conn,
		response: response,
	}
}

func (c *SMTPClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	_ = c.conn.Close()
}

func (c *SMTPClient) SendCommands(commands ...string) *SMTPResponse {
	c.lock.Lock()
	defer c.lock.Unlock()

	smtpResponse := &SMTPResponse{t: c.t}

	for _, command := range commands {
		command = strings.ReplaceAll(command, "[userAddress]", c.address)
		command = strings.ReplaceAll(command, "[userAddress|capitalize]", strings.Title(c.address))

		tstart := time.Now()

		c.debug.printReq(command)
		fmt.Fprintf(c.conn, "%s\r\n", command)

		message, err := c.response.ReadString('\n')
		if err != nil {
			smtpResponse.err = fmt.Errorf("read response failed: %v", err)
			c.debug.printErr(smtpResponse.err.Error() + "\n")
			return smtpResponse
		}

		// Message contains code and message. Codes 4xx and 5xx are bad ones, except "500 Speak up".
		if strings.HasPrefix(message, "4") || strings.HasPrefix(message, "5") {
			c.debug.printErr(message)
			err := errors.New(strings.Trim(message, "\r\n"))
			smtpResponse.err = errors.Wrap(err, "SMTP error")
			return smtpResponse
		} else if command != "" && len(message) == 0 {
			err := errors.New("empty answer")
			smtpResponse.err = errors.Wrap(err, "SMTP error")
			return smtpResponse
		}

		c.debug.printRes(message)
		smtpResponse.result = message

		c.debug.printTime(time.Since(tstart))
	}

	return smtpResponse
}

// Auth

func (c *SMTPClient) Login(account, password string) *SMTPResponse {
	c.address = account
	return c.SendCommands(
		"HELO ATEIST.TEST",
		"AUTH LOGIN",
		base64(account),
		base64(password),
	)
}

func (c *SMTPClient) Logout() *SMTPResponse {
	return c.SendCommands("QUIT")
}

// Sending

func (c *SMTPClient) EML(fileName, bcc string) *SMTPResponse {
	f, err := os.Open(fileName) //nolint:gosec
	if err != nil {
		panic(fmt.Errorf("smtp eml open: %s", err))
	}
	defer f.Close() //nolint:errcheck,gosec

	return c.SendMail(f, bcc)
}

func (c *SMTPClient) SendMail(r io.Reader, bcc string) *SMTPResponse {
	var message, from string
	var tos []string
	if bcc != "" {
		tos = append(tos, bcc)
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := string(bytes.Trim(scanner.Bytes(), "\r\n")) // Make sure no line ending is there.
		message += line + "\r\n"

		from = c.address
		if from == "" && strings.HasPrefix(line, "From: ") {
			if addrs, err := rfc5322.ParseAddressList(line[6:]); err == nil {
				from = addrs[0].Address
			}
		}
		if strings.HasPrefix(line, "To: ") || strings.HasPrefix(line, "CC: ") {
			if addrs, err := rfc5322.ParseAddressList(line[4:]); err == nil {
				for _, addr := range addrs {
					tos = append(tos, addr.Address)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("smtp eml scan: %s", err))
	}
	if from == "" {
		panic(fmt.Errorf("smtp eml no from"))
	}
	if len(tos) == 0 {
		panic(fmt.Errorf("smtp eml no to"))
	}

	commands := []string{
		fmt.Sprintf("MAIL FROM:<%s>", from),
	}
	for _, to := range tos {
		commands = append(commands, fmt.Sprintf("RCPT TO:<%s>", to))
	}
	commands = append(commands, "DATA", message+"\r\n.") // Message ending.
	return c.SendCommands(commands...)
}
