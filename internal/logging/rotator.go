// Copyright (c) 2023 Proton AG
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

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
)

type Rotator struct {
	getFile     FileProvider
	prune       Pruner
	wc          io.WriteCloser
	size        int64
	maxFileSize int64
	nextIndex   int
}

type FileProvider func(index int) (io.WriteCloser, error)

func defaultFileProvider(logsPath string, sessionID SessionID, appName AppName) FileProvider {
	return func(index int) (io.WriteCloser, error) {
		return os.Create(filepath.Join(logsPath, //nolint:gosec // G304
			fmt.Sprintf("%v_%v_%03d_v%v_%v.log", sessionID, appName, index, constants.Version, constants.Tag),
		))
	}
}

func NewRotator(maxFileSize int64, getFile FileProvider, prune Pruner) (*Rotator, error) {
	r := &Rotator{
		getFile:     getFile,
		prune:       prune,
		maxFileSize: maxFileSize,
	}

	if err := r.rotate(); err != nil {
		return nil, err
	}

	return r, nil
}

func NewDefaultRotator(logsPath string, sessionID SessionID, appName AppName, maxLogFileSize, pruningSize int64) (*Rotator, error) {
	var pruner Pruner
	if pruningSize < 0 {
		pruner = nullPruner
	} else {
		pruner = defaultPruner(logsPath, sessionID, pruningSize)
	}

	return NewRotator(maxLogFileSize, defaultFileProvider(logsPath, sessionID, appName), pruner)
}

func (r *Rotator) Write(p []byte) (int, error) {
	if r.size+int64(len(p)) > r.maxFileSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := r.wc.Write(p)
	if err != nil {
		return n, err
	}

	r.size += int64(n)
	return n, nil
}

func (r *Rotator) Close() error {
	if r.wc != nil {
		return r.wc.Close()
	}

	return nil
}

func (r *Rotator) rotate() error {
	if r.wc != nil {
		_ = r.wc.Close()
	}

	if _, err := r.prune(); err != nil {
		return err
	}

	wc, err := r.getFile(r.nextIndex)
	if err != nil {
		return err
	}

	r.nextIndex++
	r.wc = wc
	r.size = 0

	return nil
}
