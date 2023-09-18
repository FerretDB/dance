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

// Package configload provides functionality for loading and validating configuration data from YAML files.
package configload

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlekSi/pointer"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	ic "github.com/FerretDB/dance/internal/config"
)

// config represents the YAML-based configuration for the testing framework.
//
//nolint:govet // we don't care about alignment there
type config struct {
	Runner  ic.RunnerType `yaml:"runner"`
	Dir     string        `yaml:"dir"`
	Args    []string      `yaml:"args"`
	Results struct {
		Includes map[string][]string `yaml:"includes"`
		Common   *testConfig         `yaml:"common"`
		FerretDB *testConfig         `yaml:"ferretdb"`
		SQLite   *testConfig         `yaml:"sqlite"`
		MongoDB  *testConfig         `yaml:"mongodb"`
	} `yaml:"results"`
}

// testConfig represents the YAML-based configuration for database-specific test configurations.
type testConfig struct {
	Default     *ic.Status `yaml:"default"`
	Stats       *stats     `yaml:"stats"`
	Fail        []any      `yaml:"fail"`
	Skip        []any      `yaml:"skip"`
	Pass        []any      `yaml:"pass"`
	Ignore      []any      `yaml:"ignore"`
	IncludeFail []any      `yaml:"include_fail"`
}

// stats represents the YAML representation of internal config.Stats.
type stats struct {
	UnexpectedFail int `yaml:"unexpected_fail"`
	UnexpectedSkip int `yaml:"unexpected_skip"`
	UnexpectedPass int `yaml:"unexpected_pass"`
	UnexpectedRest int `yaml:"unexpected_rest"`
	ExpectedFail   int `yaml:"expected_fail"`
	ExpectedSkip   int `yaml:"expected_skip"`
	ExpectedPass   int `yaml:"expected_pass"`
}

// convertStats converts stats to internal *config.Stats.
func (s *stats) convertStats() *ic.Stats {
	if s == nil {
		return nil
	}

	return &ic.Stats{
		UnexpectedFail: s.UnexpectedFail,
		UnexpectedSkip: s.UnexpectedSkip,
		UnexpectedPass: s.UnexpectedPass,
		UnexpectedRest: s.UnexpectedRest,
		ExpectedFail:   s.ExpectedFail,
		ExpectedSkip:   s.ExpectedSkip,
		ExpectedPass:   s.ExpectedPass,
	}
}

// Load reads and validates the configuration from a YAML file,
// returning a pointer to the internal configuration struct *config.Config.
// Any error encountered during the process is also returned.
func Load(file string) (*ic.Config, error) {
	return load(file)
}

func load(file string) (*ic.Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	// Parse the YAML file into a configuration struct.
	var cfg config
	if err = d.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Validate and fill the configuration struct.
	if err = cfg.fillAndValidate(); err != nil {
		return nil, err
	}

	// Convert the YAML-based configuration to the internal representation.
	return cfg.convertAndMerge()
}

// convertAndMerge validates the YAML configuration, converts it to the internal *ic.Config,
// and merges database-specific configurations.
func (c *config) convertAndMerge() (*ic.Config, error) {
	common, err := c.Results.Common.convert(nil)
	if err != nil {
		return nil, err
	}

	includes := c.Results.Includes

	ferretDB, err := c.Results.FerretDB.convert(includes)
	if err != nil {
		return nil, err
	}

	sqLite, err := c.Results.SQLite.convert(includes)
	if err != nil {
		return nil, err
	}

	mongoDB, err := c.Results.MongoDB.convert(includes)
	if err != nil {
		return nil, err
	}

	if err := mergeCommon(common, ferretDB, sqLite, mongoDB); err != nil {
		return nil, err
	}

	return &ic.Config{
		Runner: c.Runner,
		Dir:    c.Dir,
		Args:   c.Args,
		Results: ic.Results{
			FerretDB: ferretDB,
			SQLite:   sqLite,
			MongoDB:  mongoDB,
		},
	}, nil
}

// convert converts testConfig to the internal *config.TestConfig with validation.
func (tc *testConfig) convert(includes map[string][]string) (*ic.TestConfig, error) {
	if tc == nil {
		return nil, nil
	}

	t := ic.TestConfig{
		Default: *tc.Default,
		Stats:   tc.Stats.convertStats(),
		Fail:    ic.Tests{},
		Skip:    ic.Tests{},
		Pass:    ic.Tests{},
		Ignore:  ic.Tests{},
	}

	for _, k := range tc.IncludeFail {
		includeFail := includes[k.(string)]
		t.Fail.Names = append(t.Fail.Names, includeFail...)
	}

	//nolint:govet // we don't care about alignment there
	for _, testCategory := range []struct { // testCategory examples: pass, skip sections in the yaml file
		yamlTests []any     // taken from the file, yaml representation of tests, incoming tests
		outTests  *ic.Tests // yamlTests transformed to the internal representation
	}{
		{tc.Fail, &t.Fail},
		{tc.Skip, &t.Skip},
		{tc.Pass, &t.Pass},
		{tc.Ignore, &t.Ignore},
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

	return &t, nil
}

// fillAndValidate populates the configuration with default values and performs validation.
func (c *config) fillAndValidate() error {
	// initialize common field if it's nil
	if c.Results.Common == nil {
		c.Results.Common = &testConfig{}
	}

	knownStatuses := map[ic.Status]struct{}{
		ic.Fail: {},
		ic.Skip: {},
		ic.Pass: {},
	}

	validStatus := func(status *ic.Status) bool {
		s := ic.Status(strings.ToLower(string(*status)))
		_, ok := knownStatuses[s]

		return ok
	}

	// allows us to set default values outside of common
	commonWasNil := false

	// initialize default common field if it's nil
	if c.Results.Common.Default == nil {
		c.Results.Common.Default = pointer.To(ic.Pass)
		commonWasNil = true
	}

	commonDefault := c.Results.Common.Default

	if !validStatus(commonDefault) {
		return fmt.Errorf("invalid default result: %q", *commonDefault)
	}

	for _, r := range []*testConfig{
		c.Results.FerretDB,
		c.Results.SQLite,
		c.Results.MongoDB,
	} {
		if r == nil {
			continue
		}

		// if the default value is not present use the common default and skip validation
		if r.Default == nil {
			r.Default = commonDefault
			continue
		}
		origDefault := r.Default

		if !validStatus(r.Default) {
			return fmt.Errorf("invalid default result: %q", *origDefault)
		}

		if commonWasNil && *r.Default != "" {
			continue
		}

		// this will cause a conflict so return an error
		if *commonDefault != "" && *r.Default != "" {
			return errors.New("default value cannot be set in common, when it's set in database")
		}

		// no default found so set to common default
		r.Default = commonDefault
	}

	return nil
}

// mergeCommon merges the common test configuration into database-specific test configurations
// and performs validation.
func mergeCommon(common *ic.TestConfig, configs ...*ic.TestConfig) error {
	for _, t := range configs {
		if t == nil && common == nil {
			return fmt.Errorf("all database-specific results must be set (if common results are not set)")
		}
	}

	for _, t := range configs {
		if t == nil || common == nil {
			continue
		}

		t.Fail.Names = append(t.Fail.Names, common.Fail.Names...)
		t.Skip.Names = append(t.Skip.Names, common.Skip.Names...)
		t.Pass.Names = append(t.Pass.Names, common.Pass.Names...)
		t.Ignore.Names = append(t.Ignore.Names, common.Ignore.Names...)

		t.Fail.NameRegexPattern = append(t.Fail.NameRegexPattern, common.Fail.NameRegexPattern...)
		t.Skip.NameRegexPattern = append(t.Skip.NameRegexPattern, common.Skip.NameRegexPattern...)
		t.Pass.NameRegexPattern = append(t.Pass.NameRegexPattern, common.Pass.NameRegexPattern...)
		t.Ignore.NameRegexPattern = append(t.Ignore.NameRegexPattern, common.Ignore.NameRegexPattern...)

		t.Fail.NameNotRegexPattern = append(t.Fail.NameNotRegexPattern, common.Fail.NameNotRegexPattern...)
		t.Skip.NameNotRegexPattern = append(t.Skip.NameNotRegexPattern, common.Skip.NameNotRegexPattern...)
		t.Pass.NameNotRegexPattern = append(t.Pass.NameNotRegexPattern, common.Pass.NameNotRegexPattern...)
		t.Ignore.NameNotRegexPattern = append(t.Ignore.NameNotRegexPattern, common.Ignore.NameNotRegexPattern...)

		t.Fail.OutputRegexPattern = append(t.Fail.OutputRegexPattern, common.Fail.OutputRegexPattern...)
		t.Skip.OutputRegexPattern = append(t.Skip.OutputRegexPattern, common.Skip.OutputRegexPattern...)
		t.Pass.OutputRegexPattern = append(t.Pass.OutputRegexPattern, common.Pass.OutputRegexPattern...)
		t.Ignore.OutputRegexPattern = append(t.Ignore.OutputRegexPattern, common.Ignore.OutputRegexPattern...)
	}

	for _, t := range configs {
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

	checkDuplicates := func(tc *ic.TestConfig) error {
		seen := make(map[string]struct{})

		for _, tcat := range []struct {
			tests ic.Tests
		}{
			{tc.Fail},
			{tc.Skip},
			{tc.Pass},
			{tc.Ignore},
		} {
			for _, t := range tcat.tests.Names {
				if _, ok := seen[t]; ok {
					return fmt.Errorf("duplicate test or prefix: %q", t)
				}
				seen[t] = struct{}{}
			}
		}

		return nil
	}

	for _, t := range configs {
		if err := checkDuplicates(t); err != nil {
			return err
		}
	}

	return nil
}
