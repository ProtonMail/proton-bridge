// Copyright (c) 2023 Proton AG
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
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPart(t *testing.T) {
	p := newTestParser(t, "complex_structure.eml")

	wantParts := map[string]string{
		"":        "multipart/mixed",
		"1":       "text/plain",
		"2":       "application/octet-stream",
		"3":       "multipart/mixed",
		"3.1":     "text/plain",
		"3.2":     "application/octet-stream",
		"4":       "multipart/mixed",
		"4.1":     "image/gif",
		"4.2":     "multipart/mixed",
		"4.2.1":   "text/plain",
		"4.2.2":   "multipart/alternative",
		"4.2.2.1": "text/plain",
		"4.2.2.2": "text/html",
	}

	for partNumber, wantContType := range wantParts {
		part, err := p.Section(getSectionNumber(partNumber))
		require.NoError(t, err)

		contType, _, err := part.ContentType()
		require.NoError(t, err)
		assert.Equal(t, wantContType, contType)
	}
}

func getSectionNumber(s string) (part []int) {
	if s == "" {
		return
	}

	for _, number := range strings.Split(s, ".") {
		i64, err := strconv.ParseInt(number, 10, 64)
		if err != nil {
			panic(err)
		}

		part = append(part, int(i64))
	}

	return
}

func TestPart_ConvertMetaCharset(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantErr  bool
		wantSame bool
	}{
		{
			"html no meta",
			"<body></body>",
			false,
			true,
		},
		{
			"html meta no charset",
			"<header><meta name=ProgId content=Word.Document></header><body><meta></body>",
			false,
			true,
		},
		{
			"html meta UTF-8 charset",
			"<header><meta charset=UTF-8></header><body><meta></body>",
			false,
			true,
		},
		{
			"html meta not UTF-8 charset",
			"<header><meta charset=UTF-7></header><body><meta></body>",
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p = Part{Body: []byte(tt.body)}
			err := p.ConvertMetaCharset()
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantSame, reflect.DeepEqual([]byte(tt.body), p.Body))
		})
	}
}
