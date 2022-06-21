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

package logging

import "io"

type Rotator struct {
	getFile FileProvider
	wc      io.WriteCloser
	size    int
	maxSize int
}

type FileProvider func() (io.WriteCloser, error)

func NewRotator(maxSize int, getFile FileProvider) (*Rotator, error) {
	r := &Rotator{
		getFile: getFile,
		maxSize: maxSize,
	}

	if err := r.rotate(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Rotator) Write(p []byte) (int, error) {
	if r.size+len(p) > r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := r.wc.Write(p)
	if err != nil {
		return n, err
	}

	r.size += n

	return n, nil
}

func (r *Rotator) rotate() error {
	if r.wc != nil {
		_ = r.wc.Close()
	}

	wc, err := r.getFile()
	if err != nil {
		return err
	}

	r.wc = wc
	r.size = 0

	return nil
}
