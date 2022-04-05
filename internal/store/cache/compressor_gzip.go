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

package cache

import (
	"bytes"
	"compress/gzip"
)

type GZipCompressor struct{}

func (GZipCompressor) Compress(dec []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	zw := gzip.NewWriter(buf)

	if _, err := zw.Write(dec); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (GZipCompressor) Decompress(cmp []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(cmp))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(zr); err != nil {
		return nil, err
	}

	if err := zr.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
