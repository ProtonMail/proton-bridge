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

package context

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func waitForPort(port int, timeout time.Duration) error {
	return waitUntilTrue(timeout, func() bool {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), timeout)
		if err != nil {
			return false
		}
		if conn != nil {
			if err := conn.Close(); err != nil {
				return false
			}
		}
		return true
	})
}

// waitUntilTrue can use Eventually from
// https://godoc.org/github.com/stretchr/testify/require#Assertions.Eventually
func waitUntilTrue(timeout time.Duration, callback func() bool) error {
	endTime := time.Now().Add(timeout)
	for {
		if time.Now().After(endTime) {
			return fmt.Errorf("Timeout")
		}
		if callback() {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func setLogrusVerbosityFromEnv() {
	verbosityName := os.Getenv("VERBOSITY")
	switch strings.ToLower(verbosityName) {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warning", "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.FatalLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})
}
