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

package internal

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

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

type Stats struct {
	UnexpectedRest int `yaml:"unexpected_rest"`
	UnexpectedFail int `yaml:"unexpected_fail"`
	UnexpectedSkip int `yaml:"unexpected_skip"`
	UnexpectedPass int `yaml:"unexpected_pass"`
	ExpectedFail   int `yaml:"expected_fail"`
	ExpectedSkip   int `yaml:"expected_skip"`
	ExpectedPass   int `yaml:"expected_pass"`
}

// TestsConfig represents a part of the dance configuration for tests.
//
// May contain prefixes; the longest prefix wins.
//
//nolint:govet // we don't care about alignment there
type TestsConfig struct {
	Default status   `yaml:"default"`
	Stats   Stats    `yaml:"stats"`
	Pass    []string `yaml:"pass"`
	Skip    []string `yaml:"skip"`
	Fail    []string `yaml:"fail"`
}

func (tc *TestsConfig) toMap() (map[string]status, error) {
	res := make(map[string]status, len(tc.Pass)+len(tc.Skip)+len(tc.Fail))

	for _, t := range tc.Pass {
		res[t] = Pass
	}

	for _, t := range tc.Skip {
		if _, ok := res[t]; ok {
			return nil, fmt.Errorf("duplicate test or prefix: %q", t)
		}
		res[t] = Skip
	}

	for _, t := range tc.Fail {
		if _, ok := res[t]; ok {
			return nil, fmt.Errorf("duplicate test or prefix: %q", t)
		}
		res[t] = Fail
	}

	return res, nil
}

type CompareResult struct {
	ExpectedPass   map[string]string
	ExpectedSkip   map[string]string
	ExpectedFail   map[string]string
	UnexpectedPass map[string]string
	UnexpectedSkip map[string]string
	UnexpectedFail map[string]string
	UnexpectedRest map[string]TestResult
	Stats          Stats
}

func (tc *TestsConfig) Compare(results *TestResults) (*CompareResult, error) {
	compareResult := &CompareResult{
		ExpectedPass:   make(map[string]string),
		ExpectedSkip:   make(map[string]string),
		ExpectedFail:   make(map[string]string),
		UnexpectedPass: make(map[string]string),
		UnexpectedSkip: make(map[string]string),
		UnexpectedFail: make(map[string]string),
		UnexpectedRest: make(map[string]TestResult),
	}

	tcMap, err := tc.toMap()
	if err != nil {
		return nil, err
	}

	tests := maps.Keys(results.TestResults)
	sort.Strings(tests)

	for _, test := range tests {
		testRes := results.TestResults[test]

		expectedRes := tc.Default
		for prefix := test; prefix != ""; prefix = nextPrefix(prefix) {
			if res, ok := tcMap[prefix]; ok {
				expectedRes = res
				break
			}
		}

		switch expectedRes {
		case Pass:
			switch testRes.Status {
			case Pass:
				compareResult.ExpectedPass[test] = testRes.IndentedOutput()
			case Skip:
				compareResult.UnexpectedSkip[test] = testRes.IndentedOutput()
			case Fail:
				compareResult.UnexpectedFail[test] = testRes.IndentedOutput()
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
			}
		case Skip:
			switch testRes.Status {
			case Pass:
				compareResult.UnexpectedPass[test] = testRes.IndentedOutput()
			case Skip:
				compareResult.ExpectedSkip[test] = testRes.IndentedOutput()
			case Fail:
				compareResult.UnexpectedFail[test] = testRes.IndentedOutput()
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
			}
		case Fail:
			switch testRes.Status {
			case Pass:
				compareResult.UnexpectedPass[test] = testRes.IndentedOutput()
			case Skip:
				compareResult.UnexpectedSkip[test] = testRes.IndentedOutput()
			case Fail:
				compareResult.ExpectedFail[test] = testRes.IndentedOutput()
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
			}
		case Unknown:
			fallthrough
		default:
			panic(fmt.Sprintf("unexpected expectedRes: %q", expectedRes))
		}
	}

	compareResult.Stats = Stats{
		UnexpectedRest: len(compareResult.UnexpectedRest),
		UnexpectedFail: len(compareResult.UnexpectedFail),
		UnexpectedSkip: len(compareResult.UnexpectedSkip),
		UnexpectedPass: len(compareResult.UnexpectedPass),
		ExpectedFail:   len(compareResult.ExpectedFail),
		ExpectedSkip:   len(compareResult.ExpectedSkip),
		ExpectedPass:   len(compareResult.ExpectedPass),
	}

	// special case: zero in expected_pass means "don't check"
	if tc.Stats.ExpectedPass == 0 {
		tc.Stats.ExpectedPass = compareResult.Stats.ExpectedPass
	}

	return compareResult, nil
}
