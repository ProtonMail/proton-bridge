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

package idle

import (
	"bufio"
	"errors"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/server"
)

const (
	idleCommand = "IDLE" // Capability and Command identificator
	doneLine    = "DONE"
)

// Handler for IDLE extension.
type Handler struct{}

// Command for IDLE handler.
func (h *Handler) Command() *imap.Command {
	return &imap.Command{Name: idleCommand}
}

// Parse for IDLE handler.
func (h *Handler) Parse(fields []interface{}) error {
	return nil
}

// Handle the IDLE request.
func (h *Handler) Handle(conn server.Conn) error {
	cont := &imap.ContinuationReq{Info: "idling"}
	if err := conn.WriteResp(cont); err != nil {
		return err
	}

	// Wait for DONE
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return err
	}

	if strings.ToUpper(scanner.Text()) != doneLine {
		return errors.New("expected DONE")
	}
	return nil
}

type extension struct{}

func (ext *extension) Capabilities(c server.Conn) []string {
	return []string{idleCommand}
}

func (ext *extension) Command(name string) server.HandlerFactory {
	if name != idleCommand {
		return nil
	}

	return func() server.Handler {
		return &Handler{}
	}
}

func NewExtension() server.Extension {
	return &extension{}
}
