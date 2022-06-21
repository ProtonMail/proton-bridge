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
	"bufio"
	"bytes"
	"io"
)

type boundaryReader struct {
	reader *bufio.Reader

	closed, first bool
	skipped       int

	nl               []byte // "\r\n" or "\n" (set after seeing first boundary line)
	nlDashBoundary   []byte // nl + "--boundary"
	dashBoundaryDash []byte // "--boundary--"
	dashBoundary     []byte // "--boundary"
}

func newBoundaryReader(r *bufio.Reader, boundary string) (br *boundaryReader, err error) {
	b := []byte("\r\n--" + boundary + "--")
	br = &boundaryReader{
		reader:           r,
		closed:           false,
		first:            true,
		nl:               b[:2],
		nlDashBoundary:   b[:len(b)-2],
		dashBoundaryDash: b[2:],
		dashBoundary:     b[2 : len(b)-2],
	}
	err = br.writeNextPartTo(nil)
	return
}

// writeNextPartTo will copy the the bytes of next part and write them to
// writer. Will return EOF if the underlying reader is empty.
func (br *boundaryReader) writeNextPartTo(part io.Writer) (err error) {
	if br.closed {
		return io.EOF
	}

	var line, slice []byte
	br.skipped = 0

	for {
		slice, err = br.reader.ReadSlice('\n')
		line = append(line, slice...)
		if err == bufio.ErrBufferFull {
			continue
		}

		br.skipped += len(line)

		if err == io.EOF && br.isFinalBoundary(line) {
			err = nil
			br.closed = true
			return
		}

		if err != nil {
			return
		}

		if br.isBoundaryDelimiterLine(line) {
			br.first = false
			return
		}

		if br.isFinalBoundary(line) {
			br.closed = true
			return
		}

		if part != nil {
			if _, err = part.Write(line); err != nil {
				return
			}
		}

		line = []byte{}
	}
}

func (br *boundaryReader) isFinalBoundary(line []byte) bool {
	if !bytes.HasPrefix(line, br.dashBoundaryDash) {
		return false
	}
	rest := line[len(br.dashBoundaryDash):]
	rest = skipLWSPChar(rest)
	return len(rest) == 0 || bytes.Equal(rest, br.nl)
}

func (br *boundaryReader) isBoundaryDelimiterLine(line []byte) (ret bool) {
	if !bytes.HasPrefix(line, br.dashBoundary) {
		return false
	}
	rest := line[len(br.dashBoundary):]
	rest = skipLWSPChar(rest)

	if br.first && len(rest) == 1 && rest[0] == '\n' {
		br.nl = br.nl[1:]
		br.nlDashBoundary = br.nlDashBoundary[1:]
	}
	return bytes.Equal(rest, br.nl)
}

func skipLWSPChar(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}
