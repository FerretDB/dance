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
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

// RunnerType represents the type of test runner used in the configuration.
type RunnerType string

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

// Results represents expected results for different databases.
type Results struct {
	Common   *TestsConfig
	FerretDB *TestsConfig
	MongoDB  *TestsConfig
}

// LoadConfig is used to load and validate the configuration from a YAML file.
// It reads the YAML file, decodes it, and converts it into the internal Config representation.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	var cf ConfigYAML
	if err = d.Decode(&cf); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	c, err := cf.Convert()
	if err != nil {
		return nil, err
	}

	if err = c.fillAndValidate(); err != nil {
		return nil, err
	}

	return c, nil
}

// mergeTestConfigs merges the common test configurations into database-specific test configurations.
func mergeTestConfigs(common, mongodb, ferretdb *TestsConfig) error {
	if common == nil {
		if ferretdb == nil || mongodb == nil {
			return fmt.Errorf("both FerretDB and MongoDB results must be set (if common results are not set)")
		}
		return nil
	}

	for _, t := range []struct {
		Common   *Tests
		FerretDB *Tests
		MongoDB  *Tests
	}{
		{&common.Skip, &ferretdb.Skip, &mongodb.Skip},
		{&common.Fail, &ferretdb.Fail, &mongodb.Fail},
		{&common.Pass, &ferretdb.Pass, &mongodb.Pass},
		{&common.Ignore, &ferretdb.Ignore, &mongodb.Ignore},
	} {
		t.FerretDB.Names = append(t.FerretDB.Names, t.Common.Names...)
		t.FerretDB.NameRegexPattern = append(t.FerretDB.NameRegexPattern, t.Common.NameRegexPattern...)
		t.FerretDB.NameNotRegexPattern = append(t.FerretDB.NameNotRegexPattern, t.Common.NameNotRegexPattern...)
		t.FerretDB.OutputRegexPattern = append(t.FerretDB.OutputRegexPattern, t.Common.OutputRegexPattern...)

		t.MongoDB.Names = append(t.MongoDB.Names, t.Common.Names...)
		t.MongoDB.NameRegexPattern = append(t.MongoDB.NameRegexPattern, t.Common.NameRegexPattern...)
		t.MongoDB.NameNotRegexPattern = append(t.MongoDB.NameNotRegexPattern, t.Common.NameNotRegexPattern...)
		t.MongoDB.OutputRegexPattern = append(t.MongoDB.OutputRegexPattern, t.Common.OutputRegexPattern...)
	}

	if common.Default != "" {
		if ferretdb.Default != "" || mongodb.Default != "" {
			return errors.New("default value cannot be set in common, when it's set in database")
		}
		ferretdb.Default = common.Default
		mongodb.Default = common.Default
	}

	if common.Stats != nil {
		if ferretdb.Stats != nil || mongodb.Stats != nil {
			return errors.New("stats value cannot be set in common, when it's set in database")
		}
		ferretdb.Stats = common.Stats
		mongodb.Stats = common.Stats
	}
	return nil
}

func (c *Config) fillAndValidate() error {
	if err := mergeTestConfigs(c.Results.Common, c.Results.FerretDB, c.Results.MongoDB); err != nil {
		return err
	}

	for _, r := range []*TestsConfig{
		c.Results.Common,
		c.Results.FerretDB,
		c.Results.MongoDB,
	} {
		if r == nil {
			continue
		}

		if _, err := r.toMap(); err != nil {
			return err
		}

		origDefault := r.Default
		r.Default = status(strings.ToLower(string(origDefault)))
		if r.Default == "" {
			r.Default = Pass
		}

		if _, ok := knownStatuses[r.Default]; !ok {
			return fmt.Errorf("invalid default result: %q", origDefault)
		}
	}

	return nil
}

// ForDB returns the database-specific test configuration based on the provided database name.
func (r *Results) ForDB(db string) (*TestsConfig, error) {
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

// status represents single test status.
type status string

// status values.
const (
	Pass    status = "pass"
	Skip    status = "skip"
	Fail    status = "fail"
	Ignore  status = "ignore"
	Unknown status = "unknown"
)

var knownStatuses = map[status]struct{}{
	Pass: {},
	Skip: {},
	Fail: {},
}

// TestResult represents single test result (status and output).
type TestResult struct {
	Status status
	Output string
}

func (tr *TestResult) IndentedOutput() string {
	return strings.ReplaceAll(tr.Output, "\n", "\n\t")
}

// TestResults represents results of a multiple tests.
//
// They are returned by runners.
type TestResults struct {
	// Test results by full test name.
	TestResults map[string]TestResult
}

// TestsConfig represents a part of the dance configuration for tests.
// (i.e. ferretdb/mongodb/common tests).
//
// It's used to store information about expected test results for a
// specific database.
//
// May contain prefixes; the longest prefix wins.
type TestsConfig struct {
	Default status
	Stats   *Stats
	Pass    Tests
	Skip    Tests
	Fail    Tests
	Ignore  Tests
}

// Tests are the tests from yaml category pass / fail / skip.
type Tests struct {
	Names               []string // names (i.e. "go.mongodb.org/mongo-driver/mongo/...")
	NameRegexPattern    []string // regex: "FerretDB$", the regex for the test name
	NameNotRegexPattern []string // not_regex: "FerretDB$", the regex for the test name
	OutputRegexPattern  []string // output_regex: "^server version \"5.0.9\" is (lower|higher).*"
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
	if tc.Stats.ExpectedPass == 0 {
		tc.Stats.ExpectedPass = compareResult.Stats.ExpectedPass
	}

	return compareResult, nil
}

// getExpectedStatusRegex compiles result output with expected outputs and return expected status.
// If no output matches expected - returns nil.
// If both of the regexps match, it panics.
func (tc *TestsConfig) getExpectedStatusRegex(testName string, result *TestResult) *status {
	var matchedRegex string  // name of regex that matched the test (it's required to print it on panic)
	var matchedStatus status // matched status by regex

	for _, expectedRes := range []struct {
		expectedStatus status
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

// toMap converts TestsConfig to the map of tests.
// The map stores test names as a keys and their status (pass|skip|fail), as their value.
// It returns an error if there's a test duplicate.
func (tc *TestsConfig) toMap() (map[string]status, error) {
	res := make(map[string]status, len(tc.Pass.Names)+len(tc.Skip.Names)+len(tc.Fail.Names))

	for _, tcat := range []struct {
		testsStatus status
		tests       Tests
	}{
		{Pass, tc.Pass},
		{Skip, tc.Skip},
		{Fail, tc.Fail},
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

// Structs and their internal equivalents:
// - ConfigYAML -> Config
// - ResultsYAML -> Results
// - TestsConfigYAML -> TestsConfig.

// ConfigYAML is a yaml tests representation of the Config struct.
//
// It is used only to fetch data from file. To get any of
// the dance configuration data it should be converted to
// Config struct with Convert() function.
//
//nolint:govet // we don't care about alignment there
type ConfigYAML struct {
	Runner  RunnerType  `yaml:"runner"`
	Dir     string      `yaml:"dir"`
	Args    []string    `yaml:"args"`
	Results ResultsYAML `yaml:"results"`
}

// ResultsYAML is a yaml representation of the Results struct.
type ResultsYAML struct {
	Common   *TestsConfigYAML `yaml:"common"`
	FerretDB *TestsConfigYAML `yaml:"ferretdb"`
	MongoDB  *TestsConfigYAML `yaml:"mongodb"`
}

// TestsConfigYAML is a yaml representation of the TestsConfig struct.
// It differs from it by using "any" type to be able to parse maps (i.e. "- output_regex: ...").
//
// To gain a data the struct should be first converted to TestsConfig with TestsConfigYAML.Convert() function.
type TestsConfigYAML struct {
	Default status `yaml:"default"`
	Stats   *Stats `yaml:"stats"`
	Pass    []any  `yaml:"pass"`
	Skip    []any  `yaml:"skip"`
	Fail    []any  `yaml:"fail"`
	Ignore  []any  `yaml:"ignore"`
}

// Convert validates yaml and converts ConfigYAML to the
// internal representation - Config.
func (cf *ConfigYAML) Convert() (*Config, error) {
	common, err := cf.Results.Common.Convert()
	if err != nil {
		return nil, err
	}

	ferretDB, err := cf.Results.FerretDB.Convert()
	if err != nil {
		return nil, err
	}

	mongoDB, err := cf.Results.MongoDB.Convert()
	if err != nil {
		return nil, err
	}

	return &Config{
		Runner: cf.Runner,
		Dir:    cf.Dir,
		Args:   cf.Args,
		Results: Results{
			Common:   common,
			FerretDB: ferretDB,
			MongoDB:  mongoDB,
		},
	}, nil
}

// Convert validates yaml and converts TestsConfigYAML to the
// internal representation - TestsConfig.
func (ftc *TestsConfigYAML) Convert() (*TestsConfig, error) {
	if ftc == nil {
		return nil, nil
	}

	tc := TestsConfig{ftc.Default, ftc.Stats, Tests{}, Tests{}, Tests{}, Tests{}}

	//nolint:govet // we don't care about alignment there
	for _, testCategory := range []struct { // testCategory examples: pass, skip sections in the yaml file
		yamlTests []any  // taken from the file, yaml representation of tests, incoming tests
		outTests  *Tests // yamlTests transformed to the internal representation
	}{
		{ftc.Pass, &tc.Pass},
		{ftc.Skip, &tc.Skip},
		{ftc.Fail, &tc.Fail},
		{ftc.Ignore, &tc.Ignore},
	} {
		for _, test := range testCategory.yamlTests {
			switch test := test.(type) {
			case map[string]any:
				keys := maps.Keys(test)
				if len(keys) != 1 {
					return nil, fmt.Errorf("invalid syntax: expected 1 element, got: %v", keys)
				}

				var arrPointer *[]string
				k := keys[0]
				switch k {
				case "regex":
					arrPointer = &testCategory.outTests.NameRegexPattern
				case "not_regex":
					arrPointer = &testCategory.outTests.NameNotRegexPattern
				case "output_regex":
					arrPointer = &testCategory.outTests.OutputRegexPattern
				default:
					return nil, fmt.Errorf("invalid field name %q", k)
				}

				mValue := test[k]

				regexp, ok := mValue.(string)
				if !ok {
					// Arrays are illegal:
					// - regex:
					//   - foo
					//   - bar
					if _, ok := mValue.([]string); ok {
						return nil, fmt.Errorf("invalid syntax: %s value shouldn't be an array", k)
					}
					return nil, fmt.Errorf("invalid syntax: expected string, got: %T", mValue)
				}

				// i.e. pointer to testCategory.outTests.RegexPattern = append(testCategory.outTests.RegexPattern, regexp)
				*arrPointer = append(*arrPointer, regexp)
				continue

			case string:
				testCategory.outTests.Names = append(testCategory.outTests.Names, test)
				continue

			default:
				return nil, fmt.Errorf("invalid type of %[1]q: %[1]T", test)
			}
		}
	}
	return &tc, nil
}
