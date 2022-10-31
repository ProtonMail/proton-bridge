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
	"path/filepath"
	"runtime"

	"github.com/stretchr/testify/require"
)

// Cleaner is a test step that cleans up some stuff post-test.
type Cleaner struct {
	// file is the filename of the caller.
	file string
	// lineNumber is the line number of the caller.
	lineNumber int
	// label is a descriptive label of the step being performed.
	label string
	// ctx is the TestContext on which the step operates.
	ctx *TestContext
	// cleanup is callback doing clean up.
	cleanup func()
}

// Execute runs the cleaner operation.
func (c *Cleaner) Execute() {
	c.ctx.logger.WithField("from", c.From()).Info(c.label)
	c.cleanup()
}

// From returns the filepath and line number of the place where this cleaner was created.
func (c *Cleaner) From() string {
	return fmt.Sprintf("%v:%v", c.file, c.lineNumber)
}

// addCleanup adds an operation to be performed at the end of the test.
func (ctx *TestContext) addCleanup(c func(), label string) {
	cleaner := &Cleaner{
		cleanup: c,
		label:   label,
		ctx:     ctx,
	}

	if _, file, line, ok := runtime.Caller(1); ok {
		cleaner.file, cleaner.lineNumber = filepath.Base(file), line
	}

	ctx.cleanupSteps = append([]*Cleaner{cleaner}, ctx.cleanupSteps...)
}

// addCleanupChecked adds an operation that may return an error to be performed at the end of the test.
// If the operation fails, the test is failed.
func (ctx *TestContext) addCleanupChecked(f func() error, label string) {
	checkedFunction := func() {
		err := f()
		require.NoError(ctx.t, err)
	}

	cleaner := &Cleaner{
		cleanup: checkedFunction,
		label:   label,
		ctx:     ctx,
	}

	if _, file, line, ok := runtime.Caller(1); ok {
		cleaner.file, cleaner.lineNumber = filepath.Base(file), line
	}

	ctx.cleanupSteps = append([]*Cleaner{cleaner}, ctx.cleanupSteps...)
}
