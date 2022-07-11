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
	"log"
	"regexp"
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

// Stats represent the expected/actual amount of
// failed, skipped and passed tests.
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
// TODO: purpose and cases where it's used.
//
// May contain prefixes; the longest prefix wins.
type TestsConfig struct {
	Default status
	Stats   *Stats
	Pass    Tests
	Skip    Tests
	Fail    Tests
}

// Tests are tests from yaml category pass / fail / skip
type Tests struct {
	TestNames    []string //
	RegexPattern []string // i.e. mongo.org/.*
	OutRegex     []string // does't check for names - checks output
}

// FileTestsConfig differs from TestsConfig: has any in array element
// TODO example on the the yaml to parse.
type FileTestsConfig struct {
	Default status `yaml:"default"`
	Stats   *Stats `yaml:"stats"`
	Pass    []any  `yaml:"pass"`
	Skip    []any  `yaml:"skip"`
	Fail    []any  `yaml:"fail"`
}

func (ftc *FileTestsConfig) Convert() (*TestsConfig, error) {
	if ftc == nil {
		return nil, nil // not sure if that works
	}
	tc := TestsConfig{ftc.Default, ftc.Stats, Tests{}, Tests{}, Tests{}}
	//nolint:govet // we don't care about alignment there
	for _, tcat := range []struct {
		inTests  []any
		outTests *Tests
	}{
		{ftc.Pass, &tc.Pass},
		{ftc.Skip, &tc.Skip},
		{ftc.Fail, &tc.Fail},
	} {
		for _, t := range tcat.inTests {
			switch test := t.(type) {
			case map[string]any:
				keys := maps.Keys(test)
				if len(keys) != 1 {
					return nil, fmt.Errorf("invalid syntax: expected 1 element, got: %v", keys)
				}

				var outArr *[]string
				k := keys[0]

				switch k {
				case "regex":
					outArr = &tcat.outTests.RegexPattern
				case "output_regex":
					outArr = &tcat.outTests.OutRegex
				default:
					return nil, fmt.Errorf("invalid field name: expected \"regex\" or \"output_regex\", got: %s", k)
				}

				mValue := test[k]

				regexp, ok := mValue.(string)
				if !ok {
					// Check specifically for an array
					if _, ok := mValue.([]string); ok {
						return nil, fmt.Errorf("invalid syntax: %s value shouldn't be an array", k)
					}
					return nil, fmt.Errorf("invalid syntax: expected string, got: %T", mValue)
				}

				*outArr = append(*outArr, regexp)
				//log.Fatal(outArr, tcat.outTests.NameRegex)
				continue
			case string:
				tcat.outTests.TestNames = append(tcat.outTests.TestNames, test)
				continue
			default:
				return nil, fmt.Errorf("invalid type of %[1]q: %[1]T", t)
			}
		}
	}
	return &tc, nil
}

func (tc *TestsConfig) toMap() (map[string]status, error) {
	res := make(map[string]status, len(tc.Pass.TestNames)+len(tc.Skip.TestNames)+len(tc.Fail.TestNames))

	for _, tcat := range []struct {
		testsStatus status
		tests       Tests
	}{
		{Pass, tc.Pass},
		{Skip, tc.Skip},
		{Fail, tc.Fail},
	} {
		for _, t := range tcat.tests.TestNames {
			if _, ok := res[t]; ok {
				return nil, fmt.Errorf("duplicate test or prefix: %q", t)
			}
			res[t] = tcat.testsStatus
		}
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

// Compiles result output with expected outputs and return expected status.
// If no output matches expected - returns nil.
func (tc *TestsConfig) getExpectedStatusRegex(testName string, result *TestResult) *status {
	//log.Fatal(tc.Fail.NameRegex)
	for _, expectedRes := range []struct {
		expectedStatus status
		tests          Tests
	}{
		{Pass, tc.Pass},
		{Skip, tc.Skip},
		{Fail, tc.Fail},
	} {
		// TODO: we should also check for outStatus duplicates here
		var outStatus *status
		for _, reg := range expectedRes.tests.RegexPattern {
			log.Fatal(reg)
			r := regexp.MustCompile(reg)

			if !r.MatchString(testName) {
				continue
			}
			outStatus = &expectedRes.expectedStatus
		}

		for _, reg := range expectedRes.tests.OutRegex {
			r := regexp.MustCompile(reg)

			if !r.MatchString(result.Output) {
				continue
			}
			if outStatus != nil {
				panic(fmt.Sprintf(""))
			}
			return &expectedRes.expectedStatus
		}

		return outStatus //TODO: bug
	}
	return nil
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
		expectedRes := tc.Default
		testRes := results.TestResults[test]
		//log.Fatal(tc.Fail.NameRegex)

		if expStatus := tc.getExpectedStatusRegex(test, &testRes); expStatus != nil {
			expectedRes = *expStatus
		} else {
			for prefix := test; prefix != ""; prefix = nextPrefix(prefix) {
				if res, ok := tcMap[prefix]; ok {
					expectedRes = res
					break
				}
			}
		}

		testResOutput := testRes.IndentedOutput()

		switch expectedRes {
		case Pass:
			switch testRes.Status {
			case Pass:
				compareResult.ExpectedPass[test] = testResOutput
			case Skip:
				compareResult.UnexpectedSkip[test] = testResOutput
			case Fail:
				compareResult.UnexpectedFail[test] = testResOutput
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
			}
		case Skip:
			switch testRes.Status {
			case Pass:
				compareResult.UnexpectedPass[test] = testResOutput
			case Skip:
				compareResult.ExpectedSkip[test] = testResOutput
			case Fail:
				compareResult.UnexpectedFail[test] = testResOutput
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
			}
		case Fail:
			switch testRes.Status {
			case Pass:
				compareResult.UnexpectedPass[test] = testResOutput
			case Skip:
				compareResult.UnexpectedSkip[test] = testResOutput
			case Fail:
				compareResult.ExpectedFail[test] = testResOutput
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
