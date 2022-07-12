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

// ConfigFile is a yaml tests representation of the Config struct.
//
// It is used only to fetch data from file. To get any of
// the dance configuration data it should be converted to
// Config struct with Convert() function.
//
//nolint:govet // we don't care about alignment there
type ConfigFile struct {
	Runner  string     `yaml:"runner"`
	Dir     string     `yaml:"dir"`
	Args    []string   `yaml:"args"`
	Results resultList `yaml:"results"`
}

// resultList is a yaml representation of the Results struct.
type resultList struct {
	Common   *fileTestsConfig `yaml:"common"`
	FerretDB *fileTestsConfig `yaml:"ferretdb"`
	MongoDB  *fileTestsConfig `yaml:"mongodb"`
}

// fileTestsConfig is a yaml representation of the TestsConfig struct.
// It differs from it by using "any" type to be able to parse maps (i.e. "- output_regex: ...").
//
// To gain a data the struct should be first converted to TestsConfig with fileTestsConfig.Convert() function.
type fileTestsConfig struct {
	Default status `yaml:"default"`
	Stats   *Stats `yaml:"stats"`
	Pass    []any  `yaml:"pass"`
	Skip    []any  `yaml:"skip"`
	Fail    []any  `yaml:"fail"`
}

// Converts ConfigFile to Config struct.
// ConfigFile
func (cf *ConfigFile) Convert() (*Config, error) {
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
		ResultList{common, ferretDB, mongoDB},
	}, nil
}

// Convert converts a FileTestConfig to TestConfig struct.
// FileTestsConfig it's yaml file with the tests.
// TestsConfig is an internal representation of yaml test file.
func (ftc *fileTestsConfig) Convert() (*TestsConfig, error) {
	if ftc == nil {
		return nil, nil // not sure if that works
	}

	tc := TestsConfig{ftc.Default, ftc.Stats, Tests{}, Tests{}, Tests{}}

	//nolint:govet // we don't care about alignment there
	for _, testCategory := range []struct { // testCategory examples: pass, skip sections in the yaml file
		yamlTests []any  // taken from the file, yaml representation of tests, incoming tests
		outTests  *Tests // output tests
	}{
		{ftc.Pass, &tc.Pass},
		{ftc.Skip, &tc.Skip},
		{ftc.Fail, &tc.Fail},
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
				case "output_regex":
					arrPointer = &testCategory.outTests.OutputRegexPattern
				default:
					return nil, fmt.Errorf("invalid field name: expected \"regex\" or \"output_regex\", got: %s", k)
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
