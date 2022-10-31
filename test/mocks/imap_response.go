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

package mocks

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	a "github.com/stretchr/testify/assert"
)

type IMAPResponse struct {
	t        TestingT
	err      error
	result   string
	sections []string
	done     bool
}

func (ir *IMAPResponse) sendCommand(reqTag string, reqIndex int, command string, debug *debug, conn io.Writer, response *bufio.Reader) { //nolint:interfacer
	defer func() { ir.done = true }()

	tstart := time.Now()

	commandID := fmt.Sprintf("%sO%0d", reqTag, reqIndex)
	command = fmt.Sprintf("%s %s", commandID, command)

	debug.printReq(command)
	fmt.Fprintf(conn, "%s\r\n", command)

	var section string
	for {
		line, err := response.ReadString('\n')
		if err != nil {
			ir.err = errors.Wrap(err, "read response failed")
			debug.printErr(ir.err.Error() + "\n")
			return
		}

		// Finishing line contains `commandID` following with status (`NO`, `BAD`, ...) and then message itself.
		lineWithoutID := strings.Replace(line, commandID+" ", "", 1)
		if strings.HasPrefix(line, commandID) && (strings.HasPrefix(lineWithoutID, "NO ") || strings.HasPrefix(lineWithoutID, "BAD ")) {
			debug.printErr(line)
			err := errors.New(strings.Trim(lineWithoutID, "\r\n"))
			ir.err = errors.Wrap(err, "IMAP error")
			return
		} else if command != "" && len(line) == 0 {
			err := errors.New("empty answer")
			ir.err = errors.Wrap(err, "IMAP error")
			debug.printErr(ir.err.Error() + "\n")
			return
		}
		debug.printRes(line)

		if strings.HasPrefix(line, "* ") { //nolint:gocritic
			if section != "" {
				ir.sections = append(ir.sections, section)
			}
			section = line
		} else if strings.HasPrefix(line, commandID) {
			if section != "" {
				ir.sections = append(ir.sections, section)
			}
			ir.result = line
			break
		} else {
			section += line
		}
	}

	debug.printTime(time.Since(tstart))
}

func (ir *IMAPResponse) wait() {
	for {
		if ir.done {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (ir *IMAPResponse) AssertOK() *IMAPResponse {
	ir.wait()
	a.NoError(ir.t, ir.err)
	return ir
}

func (ir *IMAPResponse) Sections() []string {
	ir.wait()
	return ir.sections
}

func (ir *IMAPResponse) AssertResult(wantResult string) *IMAPResponse {
	ir.wait()
	a.NoError(ir.t, ir.err)
	a.Regexp(ir.t, wantResult, ir.result, "Expected result %s but got %s", wantResult, ir.result)
	return ir
}

func (ir *IMAPResponse) AssertError(wantErrMsg string) *IMAPResponse {
	ir.wait()
	if ir.err == nil {
		a.Fail(ir.t, "Error is nil", "Expected to have %q", wantErrMsg)
	} else {
		a.Regexp(ir.t, wantErrMsg, ir.err.Error(), "Expected error %s but got %s", wantErrMsg, ir.err)
	}
	return ir
}

func (ir *IMAPResponse) AssertSectionsCount(expectedCount int) *IMAPResponse {
	ir.wait()
	a.Equal(ir.t, expectedCount, len(ir.sections))
	return ir
}

// AssertSectionsInOrder checks sections against regular expression in exact order.
// First regexp checks first section, second the second and so on. If there is
// more responses (sections) than expected regexps, that's OK.
func (ir *IMAPResponse) AssertSectionsInOrder(wantRegexps ...string) *IMAPResponse {
	ir.wait()
	if !a.True(ir.t,
		len(ir.sections) >= len(wantRegexps),
		"Wrong number of sections, want %v, got %v",
		len(wantRegexps),
		len(ir.sections),
	) {
		return ir
	}

	for idx, wantRegexp := range wantRegexps {
		section := ir.sections[idx]
		match, err := regexp.MatchString(wantRegexp, section)
		if !a.NoError(ir.t, err) {
			return ir
		}
		if !a.True(ir.t, match, "Section does not match given regex", section, wantRegexp) {
			return ir
		}
	}
	return ir
}

// AssertSections is similar to AssertSectionsInOrder but is not strict to the order.
// It means it just tries to find all "regexps" in the response.
func (ir *IMAPResponse) AssertSections(wantRegexps ...string) *IMAPResponse {
	ir.wait()
	for _, wantRegexp := range wantRegexps {
		a.NoError(ir.t, ir.hasSectionRegexp(wantRegexp), "regexp %v not found\nSections: %v", wantRegexp, ir.sections)
	}
	return ir
}

// AssertNotSections is similar to AssertSections but is the opposite.
// It means it just tries to find all "regexps" in the response.
func (ir *IMAPResponse) AssertNotSections(unwantedRegexps ...string) *IMAPResponse {
	ir.wait()
	for _, unwantedRegexp := range unwantedRegexps {
		a.Error(ir.t, ir.hasSectionRegexp(unwantedRegexp), "regexp %v found\nSections: %v", unwantedRegexp, ir.sections)
	}
	return ir
}

// WaitForSections is the same as AssertSections but waits for `timeout` before giving up.
func (ir *IMAPResponse) WaitForSections(timeout time.Duration, wantRegexps ...string) {
	a.Eventually(ir.t, func() bool {
		return ir.HasSections(wantRegexps...)
	}, timeout, 50*time.Millisecond, "Wanted sections: %v\nSections: %v", wantRegexps, &ir.sections)
}

// WaitForNotSections is the opposite of WaitForSection: waits to not have the response.
func (ir *IMAPResponse) WaitForNotSections(timeout time.Duration, unwantedRegexps ...string) *IMAPResponse {
	time.Sleep(timeout)
	match := ir.HasSections(unwantedRegexps...)
	a.False(ir.t, match, "Unwanted sections: %v\nSections: %v", unwantedRegexps, &ir.sections)
	return ir
}

// HasSections is the same as AssertSections but only returns bool (do not uses testingT).
func (ir *IMAPResponse) HasSections(wantRegexps ...string) bool {
	for _, wantRegexp := range wantRegexps {
		if err := ir.hasSectionRegexp(wantRegexp); err != nil {
			return false
		}
	}
	return true
}

func (ir *IMAPResponse) hasSectionRegexp(wantRegexp string) error {
	for _, section := range ir.sections {
		match, err := regexp.MatchString(wantRegexp, section)
		if err != nil {
			return err
		}
		if match {
			return nil
		}
	}
	return errors.New("Section matching given regex not found")
}
