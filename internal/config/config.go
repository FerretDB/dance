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
	"fmt"
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
// The ordering is significant and ordering improves readability, so please maintain it.
//
//nolint:musttag // we don't care about annotations there
type Stats struct {
	UnexpectedFail int
	UnexpectedSkip int
	UnexpectedPass int
	UnexpectedRest int
	ExpectedFail   int
	ExpectedSkip   int
	ExpectedPass   int
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
	PostgreSQL *TestConfig
	SQLite     *TestConfig
	MongoDB    *TestConfig
}

// TestConfig represents the configuration for tests categorized by status and regular expressions.
type TestConfig struct {
	Default Status
	Stats   *Stats
	Fail    Tests
	Skip    Tests
	Pass    Tests
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

// Tests holds information about tests of a specific status (fail, skip, pass).
type Tests struct {
	Names []string // names (i.e. "go.mongodb.org/mongo-driver/mongo/...")
}

// CompareResult encapsulates the comparison between expected and actual test outcomes.
type CompareResult struct {
	ExpectedFail map[string]string
	ExpectedSkip map[string]string
	ExpectedPass map[string]string

	UnexpectedFail map[string]string
	UnexpectedSkip map[string]string
	UnexpectedPass map[string]string

	UnexpectedRest map[string]TestResult

	Stats Stats
}

// Status represents the status of a single test.
type Status string

// Constants representing different test statuses.
const (
	Fail    Status = "fail"
	Skip    Status = "skip"
	Pass    Status = "pass"
	Ignore  Status = "ignore"  // for fluky tests
	Unknown Status = "unknown" // result can't be parsed
)

// ForDB returns the database-specific test configuration based on the provided dbName.
func (c *Config) ForDB(dbName string) (*TestConfig, error) {
	return c.Results.forDB(dbName)
}

func (r *Results) forDB(dbName string) (*TestConfig, error) {
	switch dbName {
	case "postgresql":
		if c := r.PostgreSQL; c != nil {
			return c, nil
		}
	case "sqlite":
		if c := r.SQLite; c != nil {
			return c, nil
		}
	case "mongodb":
		if c := r.MongoDB; c != nil {
			return c, nil
		}
	default:
		return nil, fmt.Errorf("unknown database %q", dbName)
	}

	return nil, fmt.Errorf("no expected results for %q", dbName)
}

// IndentedOutput returns the output of a test result with indented lines.
func (tr *TestResult) IndentedOutput() string {
	return strings.ReplaceAll(tr.Output, "\n", "\n\t")
}

// Compare compares two *TestResults and returns a *CompareResult containing the differences.
func (tc *TestConfig) Compare(results *TestResults) (*CompareResult, error) {
	compareResult := &CompareResult{
		ExpectedSkip:   make(map[string]string),
		ExpectedFail:   make(map[string]string),
		ExpectedPass:   make(map[string]string),
		UnexpectedSkip: make(map[string]string),
		UnexpectedFail: make(map[string]string),
		UnexpectedPass: make(map[string]string),
		UnexpectedRest: make(map[string]TestResult),
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
				compareResult.ExpectedSkip[test] = testResOutput
			case Pass:
				compareResult.ExpectedPass[test] = testResOutput
			case Ignore:
				fallthrough
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
			}
		case Skip:
			switch testRes.Status {
			case Fail:
				compareResult.ExpectedFail[test] = testResOutput
			case Skip:
				compareResult.ExpectedSkip[test] = testResOutput
			case Pass:
				compareResult.ExpectedPass[test] = testResOutput
			case Ignore:
				fallthrough
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
			}
		case Pass:
			switch testRes.Status {
			case Fail:
				compareResult.ExpectedFail[test] = testResOutput
			case Skip:
				compareResult.ExpectedSkip[test] = testResOutput
			case Pass:
				compareResult.ExpectedPass[test] = testResOutput
			case Ignore:
				fallthrough
			case Unknown:
				fallthrough
			default:
				compareResult.UnexpectedRest[test] = testRes
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
		UnexpectedRest: len(compareResult.UnexpectedRest),
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
