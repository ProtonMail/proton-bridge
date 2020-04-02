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

package pmapi

import (
	"errors"
	"testing"

	"github.com/getsentry/raven-go"
)

func TestSentryCrashReport(t *testing.T) {
	cm := NewClientManager(testClientConfig)
	c := cm.GetClient("bridgetest")
	if err := c.ReportSentryCrash(errors.New("Testing crash report - api proxy; goroutines with threads, find origin")); err != nil {
		t.Fatal("Expected no error while report, but have", err)
	}
}

func (s *SentryThreads) TraceAllRoutinesTest() {
	s.Values = []Thread{
		{
			ID:      0,
			Name:    "goroutine 20 [running]",
			Crashed: true,
			Stacktrace: &raven.Stacktrace{
				Frames: []*raven.StacktraceFrame{
					{
						Filename: "/home/dev/build/go-1.10.2/go/src/runtime/pprof/pprof.go",
						Function: "runtime/pprof.writeGoroutineStacks(0x9b7de0, 0xc4203e2900, 0xd0, 0xd0)",
						Lineno:   650,
					},
				},
			},
		},
		{
			ID:      1,
			Name:    "goroutine 20 [chan receive]",
			Crashed: false,
			Stacktrace: &raven.Stacktrace{
				Frames: []*raven.StacktraceFrame{
					{
						Filename: "/home/dev/build/go-1.10.2/go/src/testing/testing.go",
						Function: "testing.(*T).Run(0xc4203e42d0, 0x90f445, 0x15, 0x97d358, 0x47a501)",
						Lineno:   825,
					},
				},
			},
		},
	}
}
