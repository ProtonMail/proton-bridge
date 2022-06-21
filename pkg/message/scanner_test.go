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

package message

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	const literal = `this part of the text should be ignored

--longrandomstring

body1

--longrandomstring

body2

--longrandomstring--
`

	scanner, err := newPartScanner(strings.NewReader(literal), "longrandomstring")
	require.NoError(t, err)

	parts, err := scanner.scanAll()
	require.NoError(t, err)

	assert.Equal(t, "\nbody1\n", string(parts[0].b))
	assert.Equal(t, "\nbody2\n", string(parts[1].b))

	assert.Equal(t, "\nbody1\n", literal[parts[0].offset:parts[0].offset+len(parts[0].b)])
	assert.Equal(t, "\nbody2\n", literal[parts[1].offset:parts[1].offset+len(parts[1].b)])
}

func TestScannerNested(t *testing.T) {
	const literal = `This is the preamble.  It is to be ignored, though it 
is a handy place for mail composers to include an 
explanatory note to non-MIME compliant readers. 
--simple boundary 
Content-type: multipart/mixed; boundary="nested boundary" 

This is the preamble.  It is to be ignored, though it 
is a handy place for mail composers to include an 
explanatory note to non-MIME compliant readers. 
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--nested boundary--
--simple boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--simple boundary-- 
This is the epilogue.  It is also to be ignored.
`

	scanner, err := newPartScanner(strings.NewReader(literal), "simple boundary")
	require.NoError(t, err)

	parts, err := scanner.scanAll()
	require.NoError(t, err)

	assert.Equal(t, `Content-type: multipart/mixed; boundary="nested boundary" 

This is the preamble.  It is to be ignored, though it 
is a handy place for mail composers to include an 
explanatory note to non-MIME compliant readers. 
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--nested boundary--`, string(parts[0].b))
	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(parts[1].b))
}

func TestScannerNoFinalLinebreak(t *testing.T) {
	const literal = `--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--nested boundary--`

	scanner, err := newPartScanner(strings.NewReader(literal), "nested boundary")
	require.NoError(t, err)

	parts, err := scanner.scanAll()
	require.NoError(t, err)

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.`, string(parts[0].b))
	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(parts[1].b))
}
