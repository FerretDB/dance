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
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/AlekSi/pointer"
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
		// Includes is a mapping that allows us to merge sequences together,
		// which is currently not possible in the YAML spec - https://github.com/yaml/yaml/issues/48
		Includes   map[string][]string `yaml:"includes"`
		Common     *testConfig         `yaml:"common"`   // TODO https://github.com/FerretDB/dance/issues/591
		PostgreSQL *testConfig         `yaml:"ferretdb"` // TODO preserving YAML tag for compatibility, will update later
		SQLite     *testConfig         `yaml:"sqlite"`
		MongoDB    *testConfig         `yaml:"mongodb"`
	} `yaml:"results"`
}

// testConfig represents the YAML-based configuration for database-specific test configurations.
type testConfig struct {
	Default       *ic.Status `yaml:"default"`
	Stats         *stats     `yaml:"stats"`
	Fail          []string   `yaml:"fail"`
	Skip          []string   `yaml:"skip"`
	Pass          []string   `yaml:"pass"`
	Ignore        []string   `yaml:"ignore"`
	IncludeFail   []string   `yaml:"include_fail"`
	IncludeSkip   []string   `yaml:"include_skip"`
	IncludePass   []string   `yaml:"include_pass"`
	IncludeIgnore []string   `yaml:"include_ignore"`
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
	return cfg.convertAndValidate()
}

// convertAndValidate validates the YAML configuration, converts it to the internal *ic.Config.
func (c *config) convertAndValidate() (*ic.Config, error) {
	includes := c.Results.Includes

	postgreSQL, err := c.Results.PostgreSQL.convert(includes)
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

	if err := validate(postgreSQL, sqLite, mongoDB); err != nil {
		return nil, err
	}

	return &ic.Config{
		Runner: c.Runner,
		Dir:    c.Dir,
		Args:   c.Args,
		Results: ic.Results{
			PostgreSQL: postgreSQL,
			SQLite:     sqLite,
			MongoDB:    mongoDB,
		},
	}, nil
}

// convert converts *testConfig to the internal *ic.TestConfig.
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
		includeFail := includes[k]
		t.Fail.Names = append(t.Fail.Names, includeFail...)
	}

	for _, k := range tc.IncludePass {
		includePass := includes[k]
		t.Pass.Names = append(t.Pass.Names, includePass...)
	}

	for _, k := range tc.IncludeIgnore {
		includeIgnore := includes[k]
		t.Ignore.Names = append(t.Ignore.Names, includeIgnore...)
	}

	//nolint:govet // we don't care about alignment there
	for _, testCategory := range []struct { // testCategory examples: pass, skip sections in the yaml file
		yamlTests []string  // taken from the file, yaml representation of tests, incoming tests
		outTests  *ic.Tests // yamlTests transformed to the internal representation
	}{
		{tc.Fail, &t.Fail},
		{tc.Skip, &t.Skip},
		{tc.Pass, &t.Pass},
		{tc.Ignore, &t.Ignore},
	} {
		testCategory.outTests.Names = append(testCategory.outTests.Names, testCategory.yamlTests...)
	}

	return &t, nil
}

// fillAndValidate populates the configuration with default values and performs validation.
func (c *config) fillAndValidate() error {
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

	for _, r := range []*testConfig{
		c.Results.PostgreSQL,
		c.Results.SQLite,
		c.Results.MongoDB,
	} {
		if r == nil {
			continue
		}

		// if the default value is not present use pass
		if r.Default == nil {
			r.Default = (*ic.Status)(pointer.ToString("pass"))
			continue
		}
		origDefault := r.Default

		if !validStatus(r.Default) {
			return fmt.Errorf("invalid default result: %q", *origDefault)
		}
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
		if t == nil {
			continue
		}

		if err := checkDuplicates(t); err != nil {
			return err
		}
	}

	return nil
}
