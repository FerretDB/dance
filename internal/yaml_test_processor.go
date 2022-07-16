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

// Package internal container dance implementation.
package internal

import "fmt"

// ConfigFile is a yaml representation of the Config struct.
//
// It is used only to fetch data from file. To get any of
// the dance configuration data it should be converted to
// Config struct with Convert() function.
//
//nolint:govet // we don't care about alignment there
type ConfigFile struct {
	Runner  string      `yaml:"runner"`
	Dir     string      `yaml:"dir"`
	Args    []string    `yaml:"args"`
	Results fileResults `yaml:"results"`
}

// Convert converts a ConfigFile to Config struct.
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
		Results{common, ferretDB, mongoDB},
	}, nil
}

// fileResults is a yaml representation of the Results struct.
type fileResults struct {
	Common   *FileTestsConfig `yaml:"common"`
	FerretDB *FileTestsConfig `yaml:"ferretdb"`
	MongoDB  *FileTestsConfig `yaml:"mongodb"`
}

// FileTestsConfig is a yaml representation of the TestsConfig struct.
// It differs from it by using "any" type to be able to parse maps (i.e. "- output_regex: ...").
//
// To gain a data the struct should be first converted to TestsConfig with FileTestsConfig.Convert() function.
type FileTestsConfig struct {
	Default status `yaml:"default"`
	Stats   *Stats `yaml:"stats"`
	Pass    []any  `yaml:"pass"`
	Skip    []any  `yaml:"skip"`
	Fail    []any  `yaml:"fail"`
}

// Convert converts a FileTestsConfig to TestsConfig struct.
func (ftc *FileTestsConfig) Convert() (*TestsConfig, error) {
	if ftc == nil {
		return nil, nil // not sure if that works
	}
	tc := TestsConfig{ftc.Default, ftc.Stats, Tests{}, Tests{}, Tests{}}
	//nolint:govet // we don't care about alignment there
	for _, tcat := range []struct {
		inTests  []any
		outTests *Tests
	}{
		{ftc.Pass, &tc.Pass},
		{ftc.Skip, &tc.Skip},
		{ftc.Fail, &tc.Fail},
	} {
		for _, t := range tcat.inTests {
			switch test := t.(type) {
			case map[string]any:
				mValue, ok := test["output_regex"]
				if !ok {
					return nil, fmt.Errorf("invalid field name (\"output_regex\" expected)")
				}
				regexp, ok := mValue.(string)
				if !ok {
					// Check specifically for an array
					if _, ok := mValue.([]string); ok {
						return nil, fmt.Errorf("invalid syntax: regexp value shouldn't be an array")
					}
					return nil, fmt.Errorf("invalid syntax: expected string, got: %T", mValue)
				}

				tcat.outTests.OutRegex = append(tcat.outTests.OutRegex, regexp)
				continue
			case string:
				tcat.outTests.TestNames = append(tcat.outTests.TestNames, test)
				continue
			default:
				return nil, fmt.Errorf("invalid type of %[1]q: %[1]T", t)
			}
		}
	}
	return &tc, nil
}
