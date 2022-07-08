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

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents dance configuration.
//
//nolint:govet // we don't care about alignment there
type Config struct {
	Runner  string
	Dir     string
	Args    []string
	Results Results
}

// Results represents expected dance results.
type Results struct {
	// Expected results for both FerretDB and MongoDB.
	Common   *TestsConfig
	FerretDB *TestsConfig
	MongoDB  *TestsConfig
}

type ImportConfig struct {
	Runner  string        `yaml:"runner"`
	Dir     string        `yaml:"dir"`
	Args    []string      `yaml:"args"`
	Results importResults `yaml:"results"`
}

type importResults struct {
	Common   *ImportTestsConfig `yaml:"common"`
	FerretDB *ImportTestsConfig `yaml:"ferretdb"`
	MongoDB  *ImportTestsConfig `yaml:"mongodb"`
}

func (ic *ImportConfig) Convert() (*Config, error) {
	common, err := ic.Results.Common.Convert()
	if err != nil {
		return nil, err
	}
	ferretDB, err := ic.Results.FerretDB.Convert()
	if err != nil {
		return nil, err
	}
	mongoDB, err := ic.Results.MongoDB.Convert()
	if err != nil {
		return nil, err
	}

	return &Config{
		ic.Runner,
		ic.Dir,
		ic.Args,
		Results{common, ferretDB, mongoDB},
	}, nil
}

// Loadconfig loads and validates configuration from file.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	var ic ImportConfig
	if err = d.Decode(&ic); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	c, err := ic.Convert()
	if err != nil {
		return nil, err
	}

	if err = c.fillAndValidate(); err != nil {
		return nil, err
	}

	return c, nil
}

// mergeTestConfigs merges common config into both databases test configs.
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
	} {
		t.FerretDB.TestNames = append(t.FerretDB.TestNames, t.Common.TestNames...)
		t.FerretDB.ResRegexp = append(t.FerretDB.ResRegexp, t.Common.ResRegexp...)

		t.MongoDB.TestNames = append(t.MongoDB.TestNames, t.Common.TestNames...)
		t.MongoDB.ResRegexp = append(t.MongoDB.ResRegexp, t.Common.ResRegexp...)
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
