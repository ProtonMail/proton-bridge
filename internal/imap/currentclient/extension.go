// Copyright (c) 2020 Proton Technologies AG
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

// Package currentclient implements setting client's ID to backend.
package currentclient

import (
	"sync"
	"time"

	imapid "github.com/ProtonMail/go-imap-id"
	imapserver "github.com/emersion/go-imap/server"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "imap/currentclient") //nolint[gochecknoglobals]

type currentClientSetter interface {
	SetCurrentClient(name, version string)
}

type extension struct {
	backend currentClientSetter

	lastID       imapid.ID
	lastIDLocker sync.Locker
}

// NewExtension prepares extension reading IMAP ID from connection
// and setting it to backend. This is not standard IMAP extension.
func NewExtension(backend currentClientSetter) imapserver.Extension {
	return &extension{
		backend: backend,

		lastID:       imapid.ID{imapid.FieldName: ""},
		lastIDLocker: &sync.Mutex{},
	}
}

func (ext *extension) Capabilities(conn imapserver.Conn) []string {
	ext.readID(conn)
	return nil
}

func (ext *extension) Command(name string) imapserver.HandlerFactory {
	return nil
}

func (ext *extension) readID(conn imapserver.Conn) {
	conn.Server().ForEachConn(func(candidate imapserver.Conn) {
		if id, ok := candidate.(imapid.Conn); ok {
			if conn.Context() == candidate.Context() {
				// ID is not available right at the beginning of the connection.
				// Clients send ID quickly after AUTH. We need to wait for it.
				go func() {
					start := time.Now()
					for {
						if id.ID() != nil {
							ext.setLastID(id.ID())
							break
						}
						if time.Since(start) > 10*time.Second {
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
				}()
			}
		}
	})
}

func (ext *extension) setLastID(id imapid.ID) {
	ext.lastIDLocker.Lock()
	defer ext.lastIDLocker.Unlock()

	if name, ok := id[imapid.FieldName]; ok && ext.lastID[imapid.FieldName] != name {
		ext.lastID = imapid.ID{}
		for k, v := range id {
			ext.lastID[k] = v
		}
		log.Warn("Mail Client ID changed to ", ext.lastID)
		ext.backend.SetCurrentClient(
			ext.lastID[imapid.FieldName],
			ext.lastID[imapid.FieldVersion],
		)
	}
}
