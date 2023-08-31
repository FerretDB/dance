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

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal/config"
)

// ConfigYAML represents the YAML-based configuration for the testing framework.
//
//nolint:govet // we don't care about alignment there
type ConfigYAML struct {
	Runner  config.RunnerType `yaml:"runner"`
	Dir     string            `yaml:"dir"`
	Args    []string          `yaml:"args"`
	Results ResultsYAML       `yaml:"results"`
}

// ResultsYAML represents the YAML-based configuration for expected test results.
type ResultsYAML struct {
	Common   *TestsConfigYAML `yaml:"common"`
	FerretDB *TestsConfigYAML `yaml:"ferretdb"`
	MongoDB  *TestsConfigYAML `yaml:"mongodb"`
}

// TestsConfigYAML represents the YAML-based configuration for database-specific test configurations.
type TestsConfigYAML struct {
	Default config.Status `yaml:"default"`
	Stats   *config.Stats `yaml:"stats"`
	Pass    []any         `yaml:"pass"`
	Skip    []any         `yaml:"skip"`
	Fail    []any         `yaml:"fail"`
	Ignore  []any         `yaml:"ignore"`
}

// Load loads and validates the configuration from a YAML file.
func Load(file string) (*config.Config, error) {
	return load(file)
}

func load(file string) (*config.Config, error) {
	f, err := os.Open(file)
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

	c, err := cf.convert()
	if err != nil {
		return nil, err
	}

	if err = c.FillAndValidate(); err != nil {
		return nil, err
	}

	return c, nil
}

// convert validates the YAML configuration and converts it to the internal configuration representation.
func (cy *ConfigYAML) convert() (*config.Config, error) {
	common, err := cy.Results.Common.convert()
	if err != nil {
		return nil, err
	}

	ferretDB, err := cy.Results.FerretDB.convert()
	if err != nil {
		return nil, err
	}

	mongoDB, err := cy.Results.MongoDB.convert()
	if err != nil {
		return nil, err
	}

	return &config.Config{
		Runner: cy.Runner,
		Dir:    cy.Dir,
		Args:   cy.Args,
		Results: config.Results{
			Common:   common,
			FerretDB: ferretDB,
			MongoDB:  mongoDB,
		},
	}, nil
}

// convert converts TestsConfigYAML to the internal representation TestsConfig with validation.
func (tcy *TestsConfigYAML) convert() (*config.TestsConfig, error) {
	if tcy == nil {
		return nil, nil
	}

	tc := config.TestsConfig{
		Default: tcy.Default,
		Stats:   tcy.Stats,
		Pass:    config.Tests{},
		Skip:    config.Tests{},
		Fail:    config.Tests{},
		Ignore:  config.Tests{},
	}

	//nolint:govet // we don't care about alignment there
	for _, testCategory := range []struct { // testCategory examples: pass, skip sections in the yaml file
		yamlTests []any         // taken from the file, yaml representation of tests, incoming tests
		outTests  *config.Tests // yamlTests transformed to the internal representation
	}{
		{tcy.Pass, &tc.Pass},
		{tcy.Skip, &tc.Skip},
		{tcy.Fail, &tc.Fail},
		{tcy.Ignore, &tc.Ignore},
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
