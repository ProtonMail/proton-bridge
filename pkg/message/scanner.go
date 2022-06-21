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
	"errors"
	"io"
)

type partScanner struct {
	r *bufio.Reader

	boundary string
	progress int
}

type part struct {
	b      []byte
	offset int
}

func newPartScanner(r io.Reader, boundary string) (*partScanner, error) {
	scanner := &partScanner{r: bufio.NewReader(r), boundary: boundary}

	if _, _, err := scanner.readToBoundary(); err != nil {
		return nil, err
	}

	return scanner, nil
}

func (s *partScanner) scanAll() ([]part, error) {
	var parts []part

	for {
		offset := s.progress

		b, more, err := s.readToBoundary()
		if err != nil {
			return nil, err
		}

		if !more {
			return parts, nil
		}

		parts = append(parts, part{b: b, offset: offset})
	}
}

func (s *partScanner) readToBoundary() ([]byte, bool, error) {
	var res []byte

	for {
		line, err := s.r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, false, err
			}

			if len(line) == 0 {
				return nil, false, nil
			}
		}

		s.progress += len(line)

		switch {
		case bytes.HasPrefix(bytes.TrimSpace(line), []byte("--"+s.boundary)):
			return bytes.TrimSuffix(bytes.TrimSuffix(res, []byte("\n")), []byte("\r")), true, nil

		case bytes.HasSuffix(bytes.TrimSpace(line), []byte(s.boundary+"--")):
			return bytes.TrimSuffix(bytes.TrimSuffix(res, []byte("\n")), []byte("\r")), false, nil

		default:
			res = append(res, line...)
		}
	}
}
