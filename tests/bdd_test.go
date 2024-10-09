// Copyright (c) 2024 Proton AG
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

package tests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

type scenario struct {
	t *testCtx
}

// reset resets the test context for a new scenario.
func (s *scenario) reset(tb testing.TB) {
	s.t = newTestCtx(tb)
}

// replace replaces the placeholders in the scenario with the values from the test context.
func (s *scenario) replace(sc *godog.Scenario) {
	for _, step := range sc.Steps {
		step.Text = s.t.replace(step.Text)

		if arg := step.Argument; arg != nil {
			if table := arg.DataTable; table != nil {
				for _, row := range table.Rows {
					for _, cell := range row.Cells {
						cell.Value = s.t.replace(cell.Value)
					}
				}
			}

			if doc := arg.DocString; doc != nil {
				doc.Content = s.t.replace(doc.Content)
			}
		}
	}
}

// close closes the test context.
func (s *scenario) close(_ testing.TB) {
	s.t.close(context.Background())
}

func TestFeatures(testingT *testing.T) {
	var s scenario

	suite := godog.TestSuite{
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {
			ctx.BeforeSuite(func() {
				// Global setup.
			})

			ctx.AfterSuite(func() {
				// Global teardown.
			})
		},

		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
				s.reset(testingT)
				s.replace(sc)
				return ctx, nil
			})

			ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				s.close(testingT)
				return ctx, nil
			})

			ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
				s.t.beforeStep(st)
				return ctx, nil
			})

			ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, _ error) (context.Context, error) {
				s.t.afterStep(st, status)
				return ctx, nil
			})

			s.steps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    getFeaturePaths(),
			TestingT: testingT,
			Tags:     getFeatureTags(),
		},
	}

	if suite.Run() != 0 {
		testingT.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func getFeaturePaths() []string {
	var paths []string

	if features := os.Getenv("FEATURES"); features != "" {
		paths = strings.Split(features, " ")
	} else {
		paths = []string{"features"}
	}

	return paths
}

func getFeatureTags() string {
	var tags string

	switch arguments := os.Args; arguments[len(arguments)-1] {
	case "nightly":
		tags = "~@gmail-integration"
	case "smoke": // Currently this is just a placeholder, as there are no scenarios tagged with @smoke
		tags = "@smoke"
	case "black": // Currently this is just a placeholder, as there are no scenarios tagged with @smoke
		tags = "~@skip-black"
	case "gmail-integration":
		tags = "@gmail-integration"
	default:
		tags = "~@regression && ~@smoke && ~@gmail-integration" // To exclude more add `&& ~@tag`
	}

	return tags
}

func isBlack() bool {
	if len(os.Args) == 0 {
		return false
	}

	return os.Args[len(os.Args)-1] == "black"
}
