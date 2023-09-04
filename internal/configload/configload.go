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
		Common   *testConfig `yaml:"common"`
		FerretDB *testConfig `yaml:"ferretdb"`
		MongoDB  *testConfig `yaml:"mongodb"`
	} `yaml:"results"`
}

// testConfig represents the YAML-based configuration for database-specific test configurations.
type testConfig struct {
	Default ic.Status `yaml:"default"`
	Stats   *stats    `yaml:"stats"`
	Pass    []any     `yaml:"pass"`
	Fail    []any     `yaml:"fail"`
	Skip    []any     `yaml:"skip"`
	Ignore  []any     `yaml:"ignore"`
}

// stats represents the YAML representation of internal config.Stats.
type stats struct {
	UnexpectedRest int `yaml:"unexpected_rest"`
	UnexpectedPass int `yaml:"unexpected_pass"`
	UnexpectedFail int `yaml:"unexpected_fail"`
	UnexpectedSkip int `yaml:"unexpected_skip"`
	ExpectedPass   int `yaml:"expected_pass"`
	ExpectedFail   int `yaml:"expected_fail"`
	ExpectedSkip   int `yaml:"expected_skip"`
}

// convertStats converts stats to internal *config.Stats.
func (s *stats) convertStats() *ic.Stats {
	if s == nil {
		return nil
	}

	return &ic.Stats{
		UnexpectedRest: s.UnexpectedRest,
		UnexpectedPass: s.UnexpectedPass,
		UnexpectedFail: s.UnexpectedFail,
		UnexpectedSkip: s.UnexpectedSkip,
		ExpectedPass:   s.ExpectedPass,
		ExpectedFail:   s.ExpectedFail,
		ExpectedSkip:   s.ExpectedSkip,
	}
}

// Load loads and validates the configuration from a YAML file.
// It returns a pointer to the internal configuration struct (*ic.Config).
// If any error occurs during this process, it returns an error along with an
// error message indicating the nature of the failure.
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
	c, err := cfg.convert()
	if err != nil {
		return nil, err
	}

	// Merge specific configuration sections.
	if err := ic.MergeTestConfigs(c.Results.Common, c.Results.FerretDB, c.Results.MongoDB); err != nil {
		return nil, err
	}

	return c, nil
}

// convert validates the YAML configuration and converts it to the internal *config.Config.
func (c *config) convert() (*ic.Config, error) {
	common, err := c.Results.Common.convert()
	if err != nil {
		return nil, err
	}

	ferretDB, err := c.Results.FerretDB.convert()
	if err != nil {
		return nil, err
	}

	mongoDB, err := c.Results.MongoDB.convert()
	if err != nil {
		return nil, err
	}

	return &ic.Config{
		Runner: c.Runner,
		Dir:    c.Dir,
		Args:   c.Args,
		Results: ic.Results{
			Common:   common,
			FerretDB: ferretDB,
			MongoDB:  mongoDB,
		},
	}, nil
}

// convert converts testConfig to the internal *config.TestConfig with validation.
func (tc *testConfig) convert() (*ic.TestConfig, error) {
	if tc == nil {
		return nil, nil
	}

	t := ic.TestConfig{
		Default: tc.Default,
		Stats:   tc.Stats.convertStats(),
		Pass:    ic.Tests{},
		Fail:    ic.Tests{},
		Skip:    ic.Tests{},
		Ignore:  ic.Tests{},
	}

	//nolint:govet // we don't care about alignment there
	for _, testCategory := range []struct { // testCategory examples: pass, skip sections in the yaml file
		yamlTests []any     // taken from the file, yaml representation of tests, incoming tests
		outTests  *ic.Tests // yamlTests transformed to the internal representation
	}{
		{tc.Pass, &t.Pass},
		{tc.Fail, &t.Fail},
		{tc.Skip, &t.Skip},
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
	commonDefault := &c.Results.Common.Default

	validStatus := func(status ic.Status) bool {
		s := ic.Status(strings.ToLower(string(status)))
		_, ok := ic.KnownStatuses[s]

		return ok
	}

	if *commonDefault == "" {
		*commonDefault = ic.Pass
	}

	if !validStatus(*commonDefault) {
		return fmt.Errorf("invalid default result: %q", *commonDefault)
	}

	for _, r := range []*testConfig{
		c.Results.FerretDB,
		c.Results.MongoDB,
	} {
		if r == nil {
			continue
		}
		origDefault := &r.Default

		if *origDefault == "" {
			continue
		}

		if !validStatus(r.Default) {
			return fmt.Errorf("invalid default result: %q", *origDefault)
		}

		if *commonDefault != "" && r.Default != "" {
			return errors.New("default value cannot be set in common, when it's set in database")
		}

		// XXX this doesn't work
		r.Default = *commonDefault
	}

	return nil
}
