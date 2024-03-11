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
	ExpectedFail   int
	ExpectedSkip   int
	ExpectedPass   int
	Unknown        int
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
	PostgreSQLOldAuth *TestConfig
	PostgreSQLNewAuth *TestConfig
	SQLiteOldAuth     *TestConfig
	SQLiteNewAuth     *TestConfig
	MongoDB           *TestConfig
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

	Unknown map[string]string

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
func (c *Config) ForDB(dbName string, newAuth bool) (*TestConfig, error) {
	return c.Results.forDB(dbName, newAuth)
}

func (r *Results) forDB(dbName string, newAuth bool) (*TestConfig, error) {
	key := dbName
	if newAuth {
		key += "-newauth"
	} else {
		key += "-oldauth"
	}

	switch key {
	case "postgresql-oldauth":
		if c := r.PostgreSQLOldAuth; c != nil {
			return c, nil
		}
	case "postgresql-newauth":
		if c := r.PostgreSQLNewAuth; c != nil {
			return c, nil
		}
	case "sqlite-oldauth":
		if c := r.SQLiteOldAuth; c != nil {
			return c, nil
		}
	case "sqlite-newauth":
		if c := r.SQLiteNewAuth; c != nil {
			return c, nil
		}
	case "mongodb":
		if c := r.MongoDB; c != nil {
			return c, nil
		}
	default:
		return nil, fmt.Errorf("unknown database %q for newAuth=%t", dbName, newAuth)
	}

	return nil, fmt.Errorf("no expected results for %q and newAuth=%t", dbName, newAuth)
}

// IndentedOutput returns the output of a test result with indented lines.
func (tr *TestResult) IndentedOutput() string {
	return strings.ReplaceAll(tr.Output, "\n", "\n\t")
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
