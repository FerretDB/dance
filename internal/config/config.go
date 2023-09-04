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

// Package config provides functionality for handling and validating configuration data for test execution.
package config

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

const (
	// RunnerTypeCommand indicates a command-line test runner.
	RunnerTypeCommand RunnerType = "command"

	// RunnerTypeGoTest indicates a Go test runner.
	RunnerTypeGoTest RunnerType = "gotest"

	// RunnerTypeJSTest indicates a JavaScript test runner.
	RunnerTypeJSTest RunnerType = "jstest"

	// RunnerTypeYCSB indicates a YCSB test runner.
	RunnerTypeYCSB RunnerType = "ycsb"
)

// RunnerType represents the type of test runner used in the configuration.
type RunnerType string

// Stats represent the expected/actual amount of
// failed, skipped and passed tests.
//
//nolint:musttag // we don't care about annotations there
type Stats struct {
	UnexpectedRest int
	UnexpectedPass int
	UnexpectedFail int
	UnexpectedSkip int
	ExpectedPass   int
	ExpectedFail   int
	ExpectedSkip   int
}

// Config represents the configuration settings for the test execution.
//
//nolint:govet // we don't care about alignment there
type Config struct {
	Runner RunnerType
	Dir    string
	// Args contains additional arguments for the test runner.
	Args    []string
	Results Results
}

// Results stores the expected test results for different databases.
type Results struct {
	Common   *TestConfig
	FerretDB *TestConfig
	MongoDB  *TestConfig
}

// TestConfig represents the configuration for tests categorized by status and regular expressions.
type TestConfig struct {
	Default Status
	Stats   *Stats
	Pass    Tests
	Fail    Tests
	Skip    Tests
	Ignore  Tests
}

// TestResult represents the outcome of a single test.
type TestResult struct {
	Status Status
	Output string
}

// TestResults represents the collection of results from multiple tests.
type TestResults struct {
	// Test results by full test name.
	TestResults map[string]TestResult
}

// Tests holds information about tests of a specific status (pass, skip, fail).
type Tests struct {
	Names               []string // names (i.e. "go.mongodb.org/mongo-driver/mongo/...")
	NameRegexPattern    []string // regex: "FerretDB$", the regex for the test name
	NameNotRegexPattern []string // not_regex: "FerretDB$", the regex for the test name
	OutputRegexPattern  []string // output_regex: "^server version \"5.0.9\" is (lower|higher).*"
}

// CompareResult encapsulates the comparison between expected and actual test outcomes.
type CompareResult struct {
	ExpectedPass   map[string]string
	ExpectedFail   map[string]string
	ExpectedSkip   map[string]string
	UnexpectedPass map[string]string
	UnexpectedFail map[string]string
	UnexpectedSkip map[string]string
	UnexpectedRest map[string]TestResult
	Stats          Stats
}

// Status represents the status of a single test.
type Status string

// Constants representing different test statuses.
const (
	Pass    Status = "pass"
	Fail    Status = "fail"
	Skip    Status = "skip"
	Ignore  Status = "ignore"
	Unknown Status = "unknown"
)

// KnownStatuses is a map that contains known status values.
var KnownStatuses = map[Status]struct{}{
	Pass: {},
	Fail: {},
	Skip: {},
}

// MergeTestConfigs merges the common test configuration into database-specific test configurations.
func MergeTestConfigs(common *TestConfig, testConfig ...*TestConfig) error {
	for _, t := range testConfig {
		if t == nil && common == nil {
			return fmt.Errorf("all database-specific results must be set (if common results are not set)")
		}
	}

	for _, t := range testConfig {
		if t == nil || common == nil {
			continue
		}

		t.Pass.Names = append(t.Pass.Names, common.Pass.Names...)
		t.Fail.Names = append(t.Fail.Names, common.Fail.Names...)
		t.Skip.Names = append(t.Skip.Names, common.Skip.Names...)
		t.Ignore.Names = append(t.Ignore.Names, common.Ignore.Names...)

		t.Pass.NameRegexPattern = append(t.Pass.NameRegexPattern, common.Pass.NameRegexPattern...)
		t.Fail.NameRegexPattern = append(t.Fail.NameRegexPattern, common.Fail.NameRegexPattern...)
		t.Skip.NameRegexPattern = append(t.Skip.NameRegexPattern, common.Skip.NameRegexPattern...)
		t.Ignore.NameRegexPattern = append(t.Ignore.NameRegexPattern, common.Ignore.NameRegexPattern...)

		t.Pass.NameNotRegexPattern = append(t.Pass.NameNotRegexPattern, common.Pass.NameNotRegexPattern...)
		t.Fail.NameNotRegexPattern = append(t.Fail.NameNotRegexPattern, common.Fail.NameNotRegexPattern...)
		t.Skip.NameNotRegexPattern = append(t.Skip.NameNotRegexPattern, common.Skip.NameNotRegexPattern...)
		t.Ignore.NameNotRegexPattern = append(t.Ignore.NameNotRegexPattern, common.Ignore.NameNotRegexPattern...)

		t.Pass.OutputRegexPattern = append(t.Pass.OutputRegexPattern, common.Pass.OutputRegexPattern...)
		t.Fail.OutputRegexPattern = append(t.Fail.OutputRegexPattern, common.Fail.OutputRegexPattern...)
		t.Skip.OutputRegexPattern = append(t.Skip.OutputRegexPattern, common.Skip.OutputRegexPattern...)
		t.Ignore.OutputRegexPattern = append(t.Ignore.OutputRegexPattern, common.Ignore.OutputRegexPattern...)
	}

	for _, t := range testConfig {
		if t == nil || common == nil {
			continue
		}

		if common.Stats != nil && t.Stats != nil {
			return errors.New("stats value cannot be set in common, when it's set in database")
		}

		if common.Stats != nil {
			t.Stats = common.Stats
		}
	}

	for _, t := range testConfig {
		if _, err := t.toMap(); err != nil {
			return err
		}
	}

	return nil
}

// ForDB returns the database-specific test configuration based on the provided database name.
func (r *Results) ForDB(db string) (*TestConfig, error) {
	switch db {
	case "ferretdb":
		if c := r.FerretDB; c != nil {
			return c, nil
		}
	case "mongodb":
		if c := r.MongoDB; c != nil {
			return c, nil
		}
	default:
		return nil, fmt.Errorf("unknown database %q", db)
	}

	if c := r.Common; c != nil {
		return c, nil
	}

	return nil, fmt.Errorf("no expected results for %q", db)
}

// IndentedOutput returns the output of a test result with indented lines.
func (tr *TestResult) IndentedOutput() string {
	return strings.ReplaceAll(tr.Output, "\n", "\n\t")
}

// Compare compares two TestResults structs and returns a CompareResult containing the differences.
func (tc *TestConfig) Compare(results *TestResults) (*CompareResult, error) {
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
		case Ignore:
			continue
		case Pass:
			switch testRes.Status {
			case Pass:
				compareResult.ExpectedPass[test] = testResOutput
			case Skip:
				compareResult.UnexpectedSkip[test] = testResOutput
			case Fail:
				compareResult.UnexpectedFail[test] = testResOutput
			case Ignore:
				fallthrough
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
			case Ignore:
				fallthrough
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
			case Ignore:
				fallthrough
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
	if tc.Stats != nil && tc.Stats.ExpectedPass == 0 {
		tc.Stats.ExpectedPass = compareResult.Stats.ExpectedPass
	}

	return compareResult, nil
}

// getExpectedStatusRegex compiles result output with expected outputs and return expected status.
// If no output matches expected - returns nil.
// If both of the regexps match, it panics.
func (tc *TestConfig) getExpectedStatusRegex(testName string, result *TestResult) *Status {
	var matchedRegex string  // name of regex that matched the test (it's required to print it on panic)
	var matchedStatus Status // matched status by regex

	for _, expectedRes := range []struct {
		expectedStatus Status
		tests          Tests
	}{
		{Pass, tc.Pass},
		{Skip, tc.Skip},
		{Fail, tc.Fail},
	} {
		for _, reg := range expectedRes.tests.NameRegexPattern {
			r := regexp.MustCompile(reg)

			if !r.MatchString(testName) {
				continue
			}

			if matchedRegex != "" {
				panic(fmt.Sprintf(
					"test %s\n(output: %s)\nmatches more than one name regex: %s, %s",
					testName, result.Output, matchedRegex, reg,
				))
			}
			matchedStatus = expectedRes.expectedStatus
			matchedRegex = reg
		}

		for _, reg := range expectedRes.tests.NameNotRegexPattern {
			r := regexp.MustCompile(reg)

			if r.MatchString(testName) {
				continue
			}

			if matchedRegex != "" {
				panic(fmt.Sprintf(
					"test %s\n(output: %s)\nmatches more than one name not-regex: %s, %s",
					testName, result.Output, matchedRegex, reg,
				))
			}
			matchedStatus = expectedRes.expectedStatus
			matchedRegex = reg
		}

		for _, reg := range expectedRes.tests.OutputRegexPattern {
			r := regexp.MustCompile(reg)

			if !r.MatchString(result.Output) {
				continue
			}

			if matchedRegex != "" {
				panic(fmt.Sprintf(
					"test %s\n(output: %s)\nmatches more than one output regex: %s, %s",
					testName, result.Output, matchedRegex, reg,
				))
			}
			matchedStatus = expectedRes.expectedStatus
			matchedRegex = reg
		}
	}

	if matchedStatus == "" {
		return nil
	}

	return &matchedStatus
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

// toMap converts TestConfig to the map of tests.
// The map stores test names as a keys and their status (pass|skip|fail), as their value.
// It returns an error if there's a test duplicate.
func (tc *TestConfig) toMap() (map[string]Status, error) {
	res := make(map[string]Status, len(tc.Pass.Names)+len(tc.Skip.Names)+len(tc.Fail.Names))

	for _, tcat := range []struct {
		testsStatus Status
		tests       Tests
	}{
		{Pass, tc.Pass},
		{Fail, tc.Fail},
		{Skip, tc.Skip},
		{Ignore, tc.Ignore},
	} {
		for _, t := range tcat.tests.Names {
			if _, ok := res[t]; ok {
				return nil, fmt.Errorf("duplicate test or prefix: %q", t)
			}
			res[t] = tcat.testsStatus
		}
	}

	return res, nil
}
