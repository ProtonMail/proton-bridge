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

package message

import (
	"bufio"
	"bytes"
	"io"
	"unicode"

	"github.com/emersion/go-message/textproto"
	"github.com/pkg/errors"
)

// HeaderLines returns each line in the given header.
func HeaderLines(header []byte) [][]byte {
	var (
		lines [][]byte
		quote int
	)

	forEachLine(bufio.NewReader(bytes.NewReader(header)), func(line []byte) {
		l := bytes.SplitN(line, []byte(`: `), 2)
		isLineContinuation := quote%2 != 0 || // no quotes opened
			!bytes.Equal(bytes.TrimLeftFunc(l[0], unicode.IsSpace), l[0]) // has whitespace indent at beginning
		switch {
		case len(bytes.TrimSpace(line)) == 0:
			lines = append(lines, line)

		case isLineContinuation:
			if len(lines) > 0 {
				lines[len(lines)-1] = append(lines[len(lines)-1], line...)
			} else {
				lines = append(lines, line)
			}

		default:
			lines = append(lines, line)
		}

		quote += bytes.Count(line, []byte(`"`))
	})

	return lines
}

func forEachLine(br *bufio.Reader, fn func([]byte)) {
	for {
		b, err := br.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				panic(err)
			}

			if len(b) > 0 {
				fn(b)
			}

			return
		}

		fn(b)
	}
}

func readHeaderBody(b []byte) (*textproto.Header, []byte, error) {
	rawHeader, body, err := splitHeaderBody(b)
	if err != nil {
		return nil, nil, err
	}

	lines := HeaderLines(rawHeader)

	var header textproto.Header

	// We assume that everything before first occurrence of empty line is header.
	// If header is invalid for any reason or empty - put everything as body and let header be empty.
	if !isHeaderValid(lines) {
		return &header, b, nil
	}

	// We add lines in reverse so that calling textproto.WriteHeader later writes with the correct order.
	for i := len(lines) - 1; i >= 0; i-- {
		if len(bytes.TrimSpace(lines[i])) > 0 {
			header.AddRaw(lines[i])
		}
	}

	return &header, body, nil
}

func isHeaderValid(headerLines [][]byte) bool {
	if len(headerLines) == 0 {
		return false
	}

	for _, line := range headerLines {
		if (bytes.IndexByte(line, ':') == -1) && (len(bytes.TrimSpace(line)) > 0) {
			return false
		}
	}

	return true
}

func splitHeaderBody(b []byte) ([]byte, []byte, error) {
	br := bufio.NewReader(bytes.NewReader(b))

	var header []byte

	for {
		b, err := br.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				panic(err)
			}

			break
		}

		header = append(header, b...)

		if len(bytes.TrimSpace(b)) == 0 {
			break
		}
	}

	body, err := io.ReadAll(br)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, nil, err
	}

	return header, body, nil
}
