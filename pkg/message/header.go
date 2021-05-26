// Copyright (c) 2021 Proton Technologies AG
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

package message

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"

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
			len(l) != 2 || // it doesn't have colon
			(len(l) == 2 && !bytes.Equal(bytes.TrimSpace(l[0]), l[0])) // has white space in front of header field
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

	var header textproto.Header

	for _, line := range HeaderLines(rawHeader) {
		if len(bytes.TrimSpace(line)) > 0 {
			header.AddRaw(line)
		}
	}

	return &header, body, nil
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

	body, err := ioutil.ReadAll(br)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, nil, err
	}

	return header, body, nil
}
