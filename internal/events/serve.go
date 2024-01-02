// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package events

import "fmt"

type IMAPServerReady struct {
	eventBase

	Port int
}

func (event IMAPServerReady) String() string {
	return fmt.Sprintf("IMAPServerReady: Port %d", event.Port)
}

type IMAPServerStopped struct {
	eventBase
}

func (event IMAPServerStopped) String() string {
	return "IMAPServerStopped"
}

type IMAPServerClosed struct {
	eventBase
}

func (event IMAPServerClosed) String() string {
	return "IMAPServerClosed"
}

type IMAPServerCreated struct {
	eventBase
}

func (event IMAPServerCreated) String() string {
	return "IMAPServerCreated"
}

type IMAPServerError struct {
	eventBase

	Error error
}

func (event IMAPServerError) String() string {
	return fmt.Sprintf("IMAPServerError: %v", event.Error)
}

type SMTPServerReady struct {
	eventBase

	Port int
}

func (event SMTPServerReady) String() string {
	return fmt.Sprintf("SMTPServerReady: Port %d", event.Port)
}

type SMTPServerStopped struct {
	eventBase
}

func (event SMTPServerStopped) String() string {
	return "SMTPServerStopped"
}

type SMTPServerError struct {
	eventBase

	Error error
}

func (event SMTPServerError) String() string {
	return fmt.Sprintf("SMTPServerError: %v", event.Error)
}
