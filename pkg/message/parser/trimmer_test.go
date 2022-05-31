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

package parser

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndOfMailTrimmer(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"string without eom", "string without eom"},
		{"string with eom\r\n.\r\n", "string with eom"},
		{"string with eom\r\n.\r\nin the middle", "string with eom\r\n.\r\nin the middle"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res := dumbRead(newEndOfMailTrimmer(strings.NewReader(tt.in)))
			assert.Equal(t, tt.out, string(res))
		})
	}
}

func dumbRead(r io.Reader) []byte {
	out := []byte{}

	b := make([]byte, 1)
	for _, err := r.Read(b); err == nil; _, err = r.Read(b) {
		out = append(out, b...)
	}

	return out
}
