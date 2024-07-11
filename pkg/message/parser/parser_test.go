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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package parser

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestParser(t *testing.T, msg string) *Parser {
	p, err := New(getFileReader(msg))
	require.NoError(t, err)

	return p
}

func getFileReader(filename string) io.ReadCloser {
	f, err := os.Open(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}

	return f
}

func getFileAsString(filename string) string {
	b, err := io.ReadAll(getFileReader(filename))
	if err != nil {
		panic(err)
	}

	return string(b)
}

func FuzzNewParser(f *testing.F) {
	inSeed1, err1 := os.ReadFile(filepath.Join("testdata", "text_html_octet_attachment.eml"))
	inSeed2, err2 := os.ReadFile(filepath.Join("testdata", "complex_structure.eml"))

	require.NoError(f, err1)
	require.NoError(f, err2)

	f.Add(inSeed1)
	f.Add(inSeed2)

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = New(bytes.NewReader(data))
	})
}
