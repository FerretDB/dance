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

	"golang.org/x/exp/maps"
)

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
	Runner      string      `yaml:"runner"`
	Dir         string      `yaml:"dir"`
	Args        []string    `yaml:"args"`
	ExcludeArgs []string    `yaml:"exclude_args"`
	Results     ResultsYAML `yaml:"results"`
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
		cf.Runner,
		cf.Dir,
		cf.Args,
		cf.ExcludeArgs,
		Results{common, ferretDB, mongoDB},
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
