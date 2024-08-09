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

// Package config provides project configuration.
package config

import "fmt"

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
//nolint:vet // we don't care about alignment there
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
func (c *Config) ForDB(dbName string) (*TestConfig, error) {
	switch dbName {
	case "ferretdb-postgresql":
		if c := c.Results.PostgreSQL; c != nil {
			return c, nil
		}
	case "ferretdb-sqlite":
		if c := c.Results.SQLite; c != nil {
			return c, nil
		}
	case "mongodb":
		if c := c.Results.MongoDB; c != nil {
			return c, nil
		}
	default:
		return nil, fmt.Errorf("unknown database %q", dbName)
	}

	return nil, fmt.Errorf("no expected results for %q", dbName)
}
