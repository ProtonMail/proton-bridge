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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package message

type JobOptions struct {
	IgnoreDecryptionErrors bool // Whether to ignore decryption errors and create a "custom message" instead.
	SanitizeDate           bool // Whether to replace all dates before 1970 with RFC822's birthdate.
	AddInternalID          bool // Whether to include MessageID as X-Pm-Internal-Id.
	AddExternalID          bool // Whether to include ExternalID as X-Pm-External-Id.
	AddMessageDate         bool // Whether to include message time as X-Pm-Date.
	AddMessageIDReference  bool // Whether to include the MessageID in References.
	SanitizeMBOXHeaderLine bool // Whether to ignore header line representing MBOX delimiter
}
