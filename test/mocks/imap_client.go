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
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type IMAPClient struct {
	lock     *sync.Mutex
	debug    *debug
	t        TestingT
	reqTag   string
	reqIndex int // global request index for this client
	conn     net.Conn
	response *bufio.Reader
	idling   bool
}

func NewIMAPClient(t TestingT, tag string, imapAddr string) *IMAPClient {
	conn, err := net.Dial("tcp", imapAddr)
	require.NoError(t, err)
	if err != nil {
		return &IMAPClient{}
	}
	response := bufio.NewReader(conn)

	// Read first response to opening connection.
	_, err = response.ReadString('\n')
	assert.NoError(t, err)

	return &IMAPClient{
		lock:     &sync.Mutex{},
		debug:    newDebug(tag),
		t:        t,
		reqTag:   tag,
		reqIndex: 0,
		conn:     conn,
		response: response,
	}
}

func (c *IMAPClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.idling {
		c.StopIDLE()
	}
	_ = c.conn.Close()
}

func (c *IMAPClient) SendCommand(command string) *IMAPResponse {
	c.lock.Lock()
	defer c.lock.Unlock()

	imapResponse := &IMAPResponse{t: c.t}
	go imapResponse.sendCommand(c.reqTag, c.reqIndex, command, c.debug, c.conn, c.response)
	c.reqIndex++

	return imapResponse
}

// Auth

func (c *IMAPClient) Login(account, password string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("LOGIN %s %s", account, password))
}

func (c *IMAPClient) Logout() *IMAPResponse {
	return c.SendCommand("LOGOUT")
}

// Mailboxes

func (c *IMAPClient) ListMailboxes() *IMAPResponse {
	return c.SendCommand("LIST \"\" *")
}

func (c *IMAPClient) Select(mailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("SELECT \"%s\"", mailboxName)) //nolint:gosec
}

func (c *IMAPClient) CreateMailbox(mailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("CREATE \"%s\"", mailboxName))
}

func (c *IMAPClient) DeleteMailbox(mailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("DELETE \"%s\"", mailboxName)) //nolint:gosec
}

func (c *IMAPClient) RenameMailbox(mailboxName, newMailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("RENAME \"%s\" \"%s\"", mailboxName, newMailboxName))
}

func (c *IMAPClient) GetMailboxInfo(mailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("EXAMINE \"%s\"", mailboxName))
}

func (c *IMAPClient) GetMailboxStatus(mailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("STATUS \"%s\" (MESSAGES UNSEEN UIDNEXT UIDVALIDITY)", mailboxName))
}

// Messages

func (c *IMAPClient) FetchAllFlags() *IMAPResponse {
	return c.FetchAll("flags")
}

func (c *IMAPClient) FetchAllSubjects() *IMAPResponse {
	return c.FetchAll("body.peek[header.fields (subject)]")
}

func (c *IMAPClient) FetchAllHeaders() *IMAPResponse {
	return c.FetchAll("body.peek[header]")
}

func (c *IMAPClient) FetchAllBodyStructures() *IMAPResponse {
	return c.FetchAll("bodystructure")
}

func (c *IMAPClient) FetchAllSizes() *IMAPResponse {
	return c.FetchAll("rfc822.size")
}

func (c *IMAPClient) FetchAllBodies() *IMAPResponse {
	return c.FetchAll("rfc822")
}

func (c *IMAPClient) FetchAll(parts string) *IMAPResponse {
	return c.Fetch("1:*", parts)
}

func (c *IMAPClient) Fetch(ids, parts string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("FETCH %s %s", ids, parts))
}

func (c *IMAPClient) FetchUID(ids, parts string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("UID FETCH %s %s", ids, parts))
}

func (c *IMAPClient) Search(query string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("SEARCH %s", query))
}

// Message

func (c *IMAPClient) Append(mailboxName, msg string) *IMAPResponse {
	cmd := fmt.Sprintf("APPEND \"%s\" (\\Seen) \"%s\" {%d}\r\n%s", mailboxName, parseAppendDate(msg), len(msg), msg)
	return c.SendCommand(cmd)
}

func (c *IMAPClient) AppendBody(mailboxName, subject, from, to, body string) *IMAPResponse {
	msg := fmt.Sprintf("Subject: %s\r\n", subject)
	msg += fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Message-Id: <%d@imapbridge.com>\r\n", time.Now().Unix())
	if mailboxName != "Sent" {
		msg += "Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000\r\n"
	}
	msg += "\r\n"
	msg += body
	msg += "\r\n"

	cmd := fmt.Sprintf("APPEND \"%s\" (\\Seen) \"%s\" {%d}\r\n%s", mailboxName, parseAppendDate(msg), len(msg), msg)
	return c.SendCommand(cmd)
}

func parseAppendDate(msg string) string {
	date := "25-Mar-2021 00:30:00 +0100"
	if m, _, _, _, err := message.Parse(strings.NewReader(msg)); err == nil {
		if t, err := m.Header.Date(); err == nil {
			date = t.Format("02-Jan-2006 15:04:05 -0700")
		}
	}
	return date
}

func (c *IMAPClient) Copy(ids, newMailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("COPY %s \"%s\"", ids, newMailboxName))
}

func (c *IMAPClient) Move(ids, newMailboxName string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("MOVE %s \"%s\"", ids, newMailboxName))
}

func (c *IMAPClient) MarkAsRead(ids string) *IMAPResponse {
	return c.AddFlags(ids, "\\Seen")
}

func (c *IMAPClient) MarkAsUnread(ids string) *IMAPResponse {
	return c.RemoveFlags(ids, "\\Seen")
}

func (c *IMAPClient) MarkAsStarred(ids string) *IMAPResponse {
	return c.AddFlags(ids, "\\Flagged")
}

func (c *IMAPClient) MarkAsUnstarred(ids string) *IMAPResponse {
	return c.RemoveFlags(ids, "\\Flagged")
}

func (c *IMAPClient) MarkAsDeleted(ids string) *IMAPResponse {
	return c.AddFlags(ids, "\\Deleted")
}

func (c *IMAPClient) MarkAsUndeleted(ids string) *IMAPResponse {
	return c.RemoveFlags(ids, "\\Deleted")
}

func (c *IMAPClient) SetFlags(ids, flags string) *IMAPResponse {
	return c.changeFlags(ids, flags, "")
}

func (c *IMAPClient) AddFlags(ids, flags string) *IMAPResponse {
	return c.changeFlags(ids, flags, "+")
}

func (c *IMAPClient) RemoveFlags(ids, flags string) *IMAPResponse {
	return c.changeFlags(ids, flags, "-")
}

func (c *IMAPClient) changeFlags(ids, flags, op string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("STORE %s %sflags (%s)", ids, op, flags))
}

func (c *IMAPClient) Expunge() *IMAPResponse {
	return c.SendCommand("EXPUNGE")
}

func (c *IMAPClient) ExpungeUID(ids string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("UID EXPUNGE %s", ids))
}

func (c *IMAPClient) Noop() *IMAPResponse {
	return c.SendCommand("NOOP")
}

// Extennsions
// Extennsions: IDLE

func (c *IMAPClient) StartIDLE() *IMAPResponse {
	c.idling = true
	return c.SendCommand("IDLE")
}

func (c *IMAPClient) StopIDLE() {
	c.idling = false
	fmt.Fprintf(c.conn, "%s\r\n", "DONE")
}

// Extennsions: ID

func (c *IMAPClient) ID(request string) *IMAPResponse {
	return c.SendCommand(fmt.Sprintf("ID (%v)", request))
}
