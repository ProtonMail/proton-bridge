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

package imap

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/pkg/errors"
)

func filterHeader(header []byte, section *imap.BodySectionName) []byte {
	// Empty section.Fields means BODY[HEADER] was requested so we should return the full header.
	if len(section.Fields) == 0 {
		return header
	}

	fieldMap := make(map[string]struct{})

	for _, field := range section.Fields {
		fieldMap[strings.ToLower(field)] = struct{}{}
	}

	return filterHeaderLines(header, func(field string) bool {
		_, ok := fieldMap[strings.ToLower(field)]

		if section.NotFields {
			ok = !ok
		}

		return ok
	})
}

func filterHeaderLines(header []byte, wantField func(string) bool) []byte {
	var res []byte

	for _, line := range headerLines(header) {
		if len(bytes.TrimSpace(line)) == 0 {
			res = append(res, line...)
		} else {
			split := bytes.SplitN(line, []byte(": "), 2)

			if len(split) != 2 {
				continue
			}

			if wantField(string(bytes.ToLower(split[0]))) {
				res = append(res, line...)
			}
		}
	}

	return res
}

// NOTE: This sucks because we trim and split stuff here already, only to do it again when we use this function!
func headerLines(header []byte) [][]byte {
	var lines [][]byte

	r := bufio.NewReader(bytes.NewReader(header))

	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				panic(errors.Wrap(err, "failed to read header line"))
			}

			break
		}

		switch {
		case len(bytes.TrimSpace(b)) == 0:
			lines = append(lines, b)

		case len(bytes.SplitN(b, []byte(": "), 2)) != 2:
			lines[len(lines)-1] = append(lines[len(lines)-1], b...)

		default:
			lines = append(lines, b)
		}
	}

	return lines
}
