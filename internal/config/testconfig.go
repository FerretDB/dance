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

// TestConfig represents the configuration for tests categorized by status and regular expressions.
type TestConfig struct {
	Default Status
	Stats   *Stats
	Fail    Tests
	Skip    Tests
	Pass    Tests
	Ignore  Tests
}

// Compare compares two *TestResults and returns a *CompareResult containing the differences.
func (tc *TestConfig) Compare(results *TestResults) (*CompareResult, error) {
	compareResult := &CompareResult{
		ExpectedFail:   make(map[string]string),
		ExpectedSkip:   make(map[string]string),
		ExpectedPass:   make(map[string]string),
		UnexpectedFail: make(map[string]string),
		UnexpectedSkip: make(map[string]string),
		UnexpectedPass: make(map[string]string),
		Unknown:        make(map[string]string),
	}

	tcMap := tc.toMap()

	tests := maps.Keys(results.TestResults)
	sort.Strings(tests)

	for _, test := range tests {
		expectedRes := tc.Default
		testRes := results.TestResults[test]

		for prefix := test; prefix != ""; prefix = nextPrefix(prefix) {
			if res, ok := tcMap[prefix]; ok {
				expectedRes = res
				break
			}
		}

		testResOutput := testRes.IndentedOutput()

		switch expectedRes {
		case Fail:
			switch testRes.Status {
			case Fail:
				compareResult.ExpectedFail[test] = testResOutput
			case Skip:
				compareResult.UnexpectedSkip[test] = testResOutput
			case Pass:
				compareResult.UnexpectedPass[test] = testResOutput
			case Ignore, Unknown:
				fallthrough
			default:
				compareResult.Unknown[test] = testResOutput
			}
		case Skip:
			switch testRes.Status {
			case Fail:
				compareResult.UnexpectedFail[test] = testResOutput
			case Skip:
				compareResult.ExpectedSkip[test] = testResOutput
			case Pass:
				compareResult.UnexpectedPass[test] = testResOutput
			case Ignore, Unknown:
				fallthrough
			default:
				compareResult.Unknown[test] = testResOutput
			}
		case Pass:
			switch testRes.Status {
			case Fail:
				compareResult.UnexpectedFail[test] = testResOutput
			case Skip:
				compareResult.UnexpectedSkip[test] = testResOutput
			case Pass:
				compareResult.ExpectedPass[test] = testResOutput
			case Ignore, Unknown:
				fallthrough
			default:
				compareResult.Unknown[test] = testResOutput
			}
		case Ignore:
			continue
		case Unknown:
			fallthrough
		default:
			panic(fmt.Sprintf("unexpected expectedRes: %q", expectedRes))
		}
	}

	compareResult.Stats = Stats{
		UnexpectedFail: len(compareResult.UnexpectedFail),
		UnexpectedSkip: len(compareResult.UnexpectedSkip),
		UnexpectedPass: len(compareResult.UnexpectedPass),
		ExpectedFail:   len(compareResult.ExpectedFail),
		ExpectedSkip:   len(compareResult.ExpectedSkip),
		ExpectedPass:   len(compareResult.ExpectedPass),
		Unknown:        len(compareResult.Unknown),
	}

	return compareResult, nil
}

// toMap converts *TestConfig to the map of tests.
// The map stores test names as a keys and their status (fail|skip|pass), as their value.
func (tc *TestConfig) toMap() map[string]Status {
	res := make(map[string]Status, len(tc.Pass.Names)+len(tc.Skip.Names)+len(tc.Fail.Names))

	for _, tcat := range []struct {
		testsStatus Status
		tests       Tests
	}{
		{Fail, tc.Fail},
		{Skip, tc.Skip},
		{Pass, tc.Pass},
		{Ignore, tc.Ignore},
	} {
		for _, t := range tcat.tests.Names {
			res[t] = tcat.testsStatus
		}
	}

	return res
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
