// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
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

// Package uidplus DOES NOT implement full RFC4315!
//
// Excluded parts are:
//   - Response `UIDNOTSTICKY`: All mailboxes of Bridge support stable
//     UIDVALIDITY so it would never return this response
//
// Otherwise the standard RFC4315 is followed.
package uidplus

import (
	"errors"
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/server"
)

// Capability extension identifier.
const Capability = "UIDPLUS"

const (
	copyuid      = "COPYUID"
	appenduid    = "APPENDUID"
	copySuccess  = "COPY completed"
	appendSucess = "APPEND completed"
)

// OrderedSeq to remember Seq in order they are added.
// We didn't find any restriction in RFC that server must respond with ranges
// so we decided to always do explicit list. This makes sure that no dynamic
// ranges or out of the bound ranges are possible.
//
// NOTE: potential issue with response length
//   - the user selects large number of messages to be copied and the
//     response line will be long,
//   - list of UIDs which high values
//
// which can create long response line. We didn't find a maximum length of one
// IMAP response line or maximum length of IMAP "response code" with parameters.
type OrderedSeq []uint32

// Len return number of added seq numbers.
func (os OrderedSeq) Len() int { return len(os) }

// Add number to sequence. Zero is not acceptable UID and it won't be added to list.
func (os *OrderedSeq) Add(num uint32) {
	if num == 0 {
		return
	}
	*os = append(*os, num)
}

func (os *OrderedSeq) String() string {
	out := ""
	if len(*os) == 0 {
		return out
	}

	lastS := uint32(0)
	isRangeOpened := false
	for i, s := range *os {
		// write first
		if i == 0 {
			out += fmt.Sprintf("%d", s)
			isRangeOpened = false
			lastS = s
			continue
		}

		isLast := (i == len(*os)-1)
		isContinuous := (lastS+1 == s)

		if isContinuous {
			isRangeOpened = true
			lastS = s
			if isLast {
				out += fmt.Sprintf(":%d", s)
			}
			continue
		}

		if isRangeOpened && !isContinuous { // close range
			out += fmt.Sprintf(":%d,%d", lastS, s)
			isRangeOpened = false
			lastS = s
			continue
		}

		// Range is not opened and it is not continuous.
		out += fmt.Sprintf(",%d", s)
		isRangeOpened = false
		lastS = s
	}

	return out
}

// UIDExpunge implements server.Handler but Bridge is not supporting
// UID EXPUNGE with specific UIDs.

type UIDExpungeMailbox interface {
	Expunge() error
	UIDExpunge(*imap.SeqSet) error
}

type UIDExpunge struct {
	SeqSet *imap.SeqSet
}

func newUIDExpunge() *UIDExpunge {
	return &UIDExpunge{}
}

func (e *UIDExpunge) Parse(fields []interface{}) error {
	if len(fields) == 0 {
		return nil // It could be regular EXPUNGE without arguments.
	}
	if len(fields) > 1 {
		return errors.New("too many arguments")
	}

	seqset, ok := fields[0].(string)
	if !ok {
		return errors.New("sequence set must be an atom")
	}
	var err error
	e.SeqSet, err = imap.ParseSeqSet(seqset)
	return err
}

func (e *UIDExpunge) Handle(conn server.Conn) error {
	mailbox, err := e.getMailbox(conn)
	if err != nil {
		return err
	}
	return mailbox.Expunge()
}

func (e *UIDExpunge) UidHandle(conn server.Conn) error { //nolint:revive,stylecheck
	if e.SeqSet == nil {
		return errors.New("missing sequence set")
	}
	mailbox, err := e.getMailbox(conn)
	if err != nil {
		return err
	}
	return mailbox.UIDExpunge(e.SeqSet)
}

func (e *UIDExpunge) getMailbox(conn server.Conn) (UIDExpungeMailbox, error) {
	ctx := conn.Context()
	if ctx.Mailbox == nil {
		return nil, server.ErrNoMailboxSelected
	}
	if ctx.MailboxReadOnly {
		return nil, server.ErrMailboxReadOnly
	}

	mailbox, ok := ctx.Mailbox.(UIDExpungeMailbox)
	if !ok {
		return nil, errors.New("UID EXPUNGE is not implemented")
	}
	return mailbox, nil
}

type extension struct{}

// NewExtension of UIDPLUS.
func NewExtension() server.Extension {
	return &extension{}
}

func (ext *extension) Capabilities(c server.Conn) []string {
	if c.Context().State&imap.AuthenticatedState != 0 {
		return []string{Capability}
	}
	return nil
}

func (ext *extension) Command(name string) server.HandlerFactory {
	if name == "EXPUNGE" {
		return func() server.Handler {
			return newUIDExpunge()
		}
	}

	return nil
}

func getStatusResponseCopy(uidValidity uint32, sourceSeq, targetSeq *OrderedSeq) *imap.StatusResp {
	info := copySuccess

	if sourceSeq.Len() != 0 && targetSeq.Len() != 0 &&
		sourceSeq.Len() == targetSeq.Len() {
		info = fmt.Sprintf("[%s %d %s %s] %s",
			copyuid,
			uidValidity,
			sourceSeq.String(),
			targetSeq.String(),
			copySuccess,
		)
	}

	return &imap.StatusResp{
		Type: imap.StatusRespOk,
		Info: info,
	}
}

// CopyResponse prepares OK response with extended UID information about copied message.
func CopyResponse(uidValidity uint32, sourceSeq, targetSeq *OrderedSeq) error {
	return &imap.ErrStatusResp{
		Resp: getStatusResponseCopy(uidValidity, sourceSeq, targetSeq),
	}
}

func getStatusResponseAppend(uidValidity uint32, targetSeq *OrderedSeq) *imap.StatusResp {
	info := appendSucess
	if targetSeq.Len() > 0 {
		info = fmt.Sprintf("[%s %d %s] %s",
			appenduid,
			uidValidity,
			targetSeq.String(),
			appendSucess,
		)
	}

	return &imap.StatusResp{
		Type: imap.StatusRespOk,
		Info: info,
	}
}

// AppendResponse prepares OK response with extended UID information about appended message.
func AppendResponse(uidValidity uint32, targetSeq *OrderedSeq) error {
	return &imap.ErrStatusResp{
		Resp: getStatusResponseAppend(uidValidity, targetSeq),
	}
}
