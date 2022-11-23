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

package tests

import (
	"time"

	"github.com/ProtonMail/go-proton-api/server/backend"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
)

func init() {
	// Use the fast key generation for tests.
	backend.GenerateKey = FastGenerateKey

	// Use the fast cert generation for tests.
	certs.GenerateCert = FastGenerateCert

	// Set the event period to 100 milliseconds for more responsive tests.
	user.EventPeriod = 100 * time.Millisecond

	// Don't use jitter during tests.
	user.EventJitter = 0
}
