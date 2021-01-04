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

package id

import (
	imapid "github.com/ProtonMail/go-imap-id"
	imapserver "github.com/emersion/go-imap/server"
)

type currentClientSetter interface {
	SetCurrentClient(name, version string)
}

// Extension for IMAP server
type extension struct {
	extID  imapserver.ConnExtension
	setter currentClientSetter
}

func (ext *extension) Capabilities(conn imapserver.Conn) []string {
	return ext.extID.Capabilities(conn)
}

func (ext *extension) Command(name string) imapserver.HandlerFactory {
	newIDHandler := ext.extID.Command(name)
	if newIDHandler == nil {
		return nil
	}
	return func() imapserver.Handler {
		if hdlrID, ok := newIDHandler().(*imapid.Handler); ok {
			return &handler{
				hdlrID: hdlrID,
				setter: ext.setter,
			}
		}
		return nil
	}
}

func (ext *extension) NewConn(conn imapserver.Conn) imapserver.Conn {
	return ext.extID.NewConn(conn)
}

type handler struct {
	hdlrID *imapid.Handler
	setter currentClientSetter
}

func (hdlr *handler) Parse(fields []interface{}) error {
	return hdlr.hdlrID.Parse(fields)
}

func (hdlr *handler) Handle(conn imapserver.Conn) error {
	err := hdlr.hdlrID.Handle(conn)
	if err == nil {
		id := hdlr.hdlrID.Command.ID
		hdlr.setter.SetCurrentClient(
			id[imapid.FieldName],
			id[imapid.FieldVersion],
		)
	}
	return err
}

// NewExtension returns extension which is adding RFC2871 ID capability, with
// direct interface to set information about email client to backend.
func NewExtension(serverID imapid.ID, setter currentClientSetter) imapserver.Extension {
	if conExtID, ok := imapid.NewExtension(serverID).(imapserver.ConnExtension); ok {
		return &extension{
			extID:  conExtID,
			setter: setter,
		}
	}
	return nil
}
