// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// TestResult represents the actual outcome of a single test.
//
//nolint:vet // for readability
type TestResult struct {
	Status       Status
	Output       string
	Measurements map[string]float64
}

// IndentedOutput returns the output of a test result with indented lines.
func (tr *TestResult) IndentedOutput() string {
	return strings.ReplaceAll(tr.Output, "\n", "\n\t")
}

// CompareResults represents the comparison between expected and actual test outcomes.
type CompareResults struct {
	// expected
	Failed  map[string]TestResult
	Skipped map[string]TestResult
	Passed  map[string]TestResult

	// unexpected
	XFailed  map[string]TestResult
	XSkipped map[string]TestResult
	XPassed  map[string]TestResult

	Unknown map[string]TestResult

	Stats Stats
}

// ExpectedResults represents expected results for specific database.
type ExpectedResults struct {
	Default Status
	Stats   *Stats

	// test names
	Fail   []string
	Skip   []string
	Pass   []string
	Ignore []string
}

func (expected *ExpectedResults) mapStatuses() map[string]Status {
	res := make(map[string]Status)

	for _, g := range []struct {
		status Status
		names  []string
	}{
		{Fail, expected.Fail},
		{Skip, expected.Skip},
		{Pass, expected.Pass},
		{Ignore, expected.Ignore},
	} {
		for _, n := range g.names {
			res[n] = g.status
		}
	}

	return res
}

// Compare compares expected and actual results.
func (expected *ExpectedResults) Compare(actual map[string]TestResult) (*CompareResults, error) {
	res := &CompareResults{
		Failed:   make(map[string]TestResult),
		Skipped:  make(map[string]TestResult),
		Passed:   make(map[string]TestResult),
		XFailed:  make(map[string]TestResult),
		XSkipped: make(map[string]TestResult),
		XPassed:  make(map[string]TestResult),
		Unknown:  make(map[string]TestResult),
	}

	tests := maps.Keys(actual)
	sort.Strings(tests)

	m := expected.mapStatuses()

	for _, test := range tests {
		actualResult := actual[test]

		expectedStatus := expected.Default

		for prefix := test; prefix != ""; prefix = nextPrefix(prefix) {
			if res, ok := m[prefix]; ok {
				expectedStatus = res
				break
			}
		}

		o := actualResult.IndentedOutput()
		tr := TestResult{
			Status:       actualResult.Status,
			Output:       o,
			Measurements: actualResult.Measurements,
		}

		switch expectedStatus {
		case Fail:
			switch actualResult.Status {
			case Fail:
				res.Failed[test] = tr
			case Skip:
				res.XSkipped[test] = tr
			case Pass:
				res.XPassed[test] = tr
			case Ignore, Unknown:
				fallthrough
			default:
				res.Unknown[test] = tr
			}

		case Skip:
			switch actualResult.Status {
			case Fail:
				res.XFailed[test] = tr
			case Skip:
				res.Skipped[test] = tr
			case Pass:
				res.XPassed[test] = tr
			case Ignore, Unknown:
				fallthrough
			default:
				res.Unknown[test] = tr
			}

		case Pass:
			switch actualResult.Status {
			case Fail:
				res.XFailed[test] = tr
			case Skip:
				res.XSkipped[test] = tr
			case Pass:
				res.Passed[test] = tr
			case Ignore, Unknown:
				fallthrough
			default:
				res.Unknown[test] = tr
			}

		case Ignore:
			continue
		case Unknown:
			fallthrough
		default:
			panic(fmt.Sprintf("unexpected expectedStatus %q", expectedStatus))
		}
	}

	res.Stats = Stats{
		Failed:   len(res.Failed),
		Skipped:  len(res.Skipped),
		Passed:   len(res.Passed),
		XFailed:  len(res.XFailed),
		XSkipped: len(res.XSkipped),
		XPassed:  len(res.XPassed),
		Unknown:  len(res.Unknown),
	}

	return res, nil
}

// nextPrefix returns the next prefix of the given path, stopping on / and .
// It panics for empty string.
func nextPrefix(path string) string {
	if path == "" {
		panic("path is empty")
	}

	if t := strings.TrimRight(path, "."); t != path {
		return t
	}

	if t := strings.TrimRight(path, "/"); t != path {
		return t
	}

	i := strings.LastIndexAny(path, "/.")

	return path[:i+1]
}
