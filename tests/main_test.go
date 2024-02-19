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

package tests

import (
	"os"
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api/server/backend"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/sirupsen/logrus"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	// Use the fast key generation for tests.
	backend.GenerateKey = backend.FastGenerateKey

	// Use the fast cert generation for tests.
	certs.GenerateCert = FastGenerateCert

	if !isBlack() {
		// Set the event period to 1 second for more responsive tests.
		user.EventPeriod = time.Second
		// Don't use jitter during tests.
		user.EventJitter = 0
	}

	level := os.Getenv("FEATURE_TEST_LOG_LEVEL")

	if os.Getenv("BRIDGE_API_DEBUG") != "" {
		level = "trace"
	}

	if level != "" {
		if parsed, err := logrus.ParseLevel(level); err == nil {
			logrus.SetLevel(parsed)
		}
	}

	goleak.VerifyTestMain(m, goleak.IgnoreCurrent())
}
