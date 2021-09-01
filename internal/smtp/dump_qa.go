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

// +build build_qa

package smtp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func dumpMessageData(b []byte, subject string) {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.WithError(err).Error("Failed to dump raw message data")
		return
	}

	path := filepath.Join(home, "bridge-qa")

	if err := os.MkdirAll(path, 0700); err != nil {
		logrus.WithError(err).Error("Failed to dump raw message data")
		return
	}

	if len(subject) > 16 {
		subject = subject[:16]
	}

	if err := ioutil.WriteFile(
		filepath.Join(path, fmt.Sprintf("%v-%v.eml", subject, time.Now().Unix())),
		b,
		0600,
	); err != nil {
		logrus.WithError(err).Error("Failed to dump raw message data")
		return
	}

	logrus.WithField("path", path).Info("Dumped raw message data")
}
