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

// Package files provides standard filesystem operations.
package files

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type OpRemove struct {
	targets    []string
	exceptions []string
}

func Remove(targets ...string) *OpRemove {
	return &OpRemove{targets: targets}
}

func (op *OpRemove) Except(exceptions ...string) *OpRemove {
	op.exceptions = exceptions
	return op
}

func (op *OpRemove) Do() error {
	var multiErr error

	for _, target := range op.targets {
		if err := remove(target, op.exceptions...); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr
}

func remove(dir string, except ...string) error {
	var toRemove []string

	if err := filepath.Walk(dir, func(path string, _ os.FileInfo, _ error) error {
		for _, exception := range except {
			if path == exception || strings.HasPrefix(exception, path) || strings.HasPrefix(path, exception) {
				return nil
			}
		}

		toRemove = append(toRemove, path)

		return nil
	}); err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(toRemove)))

	var multiErr error
	for _, target := range toRemove {
		if err := os.RemoveAll(target); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr
}
