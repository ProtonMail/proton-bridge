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

package parser

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestParser(t *testing.T, msg string) *Parser {
	r := f(msg)

	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(r); err != nil {
		panic(err)
	}

	p, err := New(buf.Bytes())
	require.NoError(t, err)

	return p
}

func f(filename string) io.ReadCloser {
	f, err := os.Open(filepath.Join("testdata", filename))

	if err != nil {
		panic(err)
	}

	return f
}

func s(filename string) string {
	b, err := ioutil.ReadAll(f(filename))
	if err != nil {
		panic(err)
	}

	return string(b)
}
