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
	"fmt"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/getsentry/raven-go"
)

const fileParseError = "[file parse error]"

var isGoroutine = regexp.MustCompile("^goroutine [[:digit:]]+.*") //nolint[gochecknoglobals]

// SentryThreads implements standard sentry thread report.
type SentryThreads struct {
	Values []Thread `json:"values"`
}

// Class specifier.
func (s *SentryThreads) Class() string { return "threads" }

// Thread wraps a single stacktrace.
type Thread struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Crashed    bool              `json:"crashed"`
	Stacktrace *raven.Stacktrace `json:"stacktrace"`
}

// TraceAllRoutines traces all goroutines and saves them to the current object.
func (s *SentryThreads) TraceAllRoutines() {
	s.Values = []Thread{}
	goroutines := &strings.Builder{}
	_ = pprof.Lookup("goroutine").WriteTo(goroutines, 2)

	thread := Thread{ID: -1}
	var frame *raven.StacktraceFrame
	for _, v := range strings.Split(goroutines.String(), "\n") {
		// Ignore empty lines.
		if v == "" {
			continue
		}

		// New routine.
		if isGoroutine.MatchString(v) {
			if thread.ID >= 0 {
				s.Values = append(s.Values, thread)
			}
			thread = Thread{ID: thread.ID + 1, Name: v, Crashed: thread.ID == -1, Stacktrace: &raven.Stacktrace{Frames: []*raven.StacktraceFrame{}}}
			continue
		}

		// New function.
		if frame == nil {
			frame = &raven.StacktraceFrame{Function: v}
			continue
		}

		// Set filename and add frame.
		if frame.Filename == "" {
			fld := strings.Fields(v)
			if len(fld) != 2 {
				frame.Filename = fileParseError
				frame.AbsolutePath = v
			} else {
				frame.Filename = fld[0]
				sp := strings.Split(fld[0], ":")
				if len(sp) > 1 {
					i, err := strconv.Atoi(sp[len(sp)-1])
					if err == nil {
						frame.Filename = strings.Join(sp[:len(sp)-1], ":")
						frame.Lineno = i
					}
				}
			}
			if frame.AbsolutePath == "" && frame.Filename != fileParseError {
				frame.AbsolutePath = frame.Filename
				if sp := strings.Split(frame.Filename, "/"); len(sp) > 1 {
					frame.Filename = sp[len(sp)-1]
				}
			}
			thread.Stacktrace.Frames = append([]*raven.StacktraceFrame{frame}, thread.Stacktrace.Frames...)
			frame = nil
			continue
		}
	}
	// Add last thread.
	s.Values = append(s.Values, thread)
}

func findPanicSender(s *SentryThreads, err error) string {
	out := "error nil"
	if err != nil {
		out = err.Error()
	}
	for _, thread := range s.Values {
		if !thread.Crashed {
			continue
		}
		for i, fr := range thread.Stacktrace.Frames {
			if strings.HasSuffix(fr.Filename, "panic.go") && strings.HasPrefix(fr.Function, "panic") {
				// Next frame if any.
				j := 0
				if i > j {
					j = i - 1
				}

				// Directory and filename.
				fname := thread.Stacktrace.Frames[j].AbsolutePath
				if sp := strings.Split(fname, "/"); len(sp) > 2 {
					fname = strings.Join(sp[len(sp)-2:], "/")
				}

				// Line number.
				if ln := thread.Stacktrace.Frames[j].Lineno; ln > 0 {
					fname = fmt.Sprintf("%s:%d", fname, ln)
				}

				out = fmt.Sprintf("%s: %s", fname, out)
				break // Just first panic.
			}
		}
	}
	return out
}

// ReportSentryCrash reports a sentry crash with stacktrace from all goroutines.
func (c *Client) ReportSentryCrash(reportErr error) (err error) {
	if reportErr == nil {
		return
	}
	tags := map[string]string{
		"OS":        runtime.GOOS,
		"Client":    c.cm.GetConfig().ClientID,
		"Version":   c.cm.GetConfig().AppVersion,
		"UserAgent": CurrentUserAgent,
		"UserID":    c.userID,
	}

	threads := &SentryThreads{}
	threads.TraceAllRoutines()
	errorWithFile := findPanicSender(threads, reportErr)
	packet := raven.NewPacket(errorWithFile, threads)

	eventID, ch := raven.Capture(packet, tags)
	if err = <-ch; err == nil {
		c.log.Warn("Reported error with id: ", eventID)
	} else {
		c.log.Errorf("Can not report `%s` due to `%s`", reportErr.Error(), err.Error())
	}
	return err
}
