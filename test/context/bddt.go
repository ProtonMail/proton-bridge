// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type bddT struct {
	err *multierror.Error
}

func (t *bddT) Errorf(msg string, args ...interface{}) {
	err := fmt.Errorf(msg, args...)
	t.err = multierror.Append(t.err, err)
}

func (t *bddT) FailNow() {
	t.err = multierror.Append(t.err, errors.New("failed by calling FailNow"))
}

func (t *bddT) getErrors() error {
	return t.err.ErrorOrNil()
}
