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

package imap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterHeader(t *testing.T) {
	const header = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\n\r\n"

	assert.Equal(t, "To: somebody\r\n\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "To")
	})))

	assert.Equal(t, "From: somebody else\r\n\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "From")
	})))

	assert.Equal(t, "To: somebody\r\nFrom: somebody else\r\n\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "To") || strings.EqualFold(field, "From")
	})))

	assert.Equal(t, "Subject: this is\r\n\ta multiline field\r\n\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "Subject")
	})))
}

// TestFilterHeaderNoNewline tests that we don't include a trailing newline when filtering
// if the original header also lacks one (which it can legally do if there is no body).
func TestFilterHeaderNoNewline(t *testing.T) {
	const header = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\n"

	assert.Equal(t, "To: somebody\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "To")
	})))

	assert.Equal(t, "From: somebody else\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "From")
	})))

	assert.Equal(t, "To: somebody\r\nFrom: somebody else\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "To") || strings.EqualFold(field, "From")
	})))

	assert.Equal(t, "Subject: this is\r\n\ta multiline field\r\n", string(filterHeaderLines([]byte(header), func(field string) bool {
		return strings.EqualFold(field, "Subject")
	})))
}
