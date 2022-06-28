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
	Runner  string   `yaml:"runner"`
	Dir     string   `yaml:"dir"`
	Args    []string `yaml:"args"`
	Results Results  `yaml:"results"`
}

// Results represents expected dance results.
type Results struct {
	// Expected results for both FerretDB and MongoDB.
	Common   *TestsConfig `yaml:"common"`
	FerretDB *TestsConfig `yaml:"ferretdb"`
	MongoDB  *TestsConfig `yaml:"mongodb"`
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

	var c Config
	if err = d.Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if err = c.fillAndValidate(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) fillAndValidate() error {
	if c.Results.Common != nil {
		for _, result := range c.Results.Common.Skip {
			c.Results.MongoDB.Skip = append(c.Results.MongoDB.Skip, result)
			c.Results.FerretDB.Skip = append(c.Results.FerretDB.Skip, result)
		}
		for _, result := range c.Results.Common.Fail {
			c.Results.MongoDB.Fail = append(c.Results.MongoDB.Fail, result)
			c.Results.FerretDB.Fail = append(c.Results.FerretDB.Fail, result)
		}
		for _, result := range c.Results.Common.Pass {
			c.Results.MongoDB.Pass = append(c.Results.MongoDB.Pass, result)
			c.Results.FerretDB.Pass = append(c.Results.FerretDB.Pass, result)
		}

		if c.Results.Common.Default != "" {
			if c.Results.FerretDB.Default != "" || c.Results.MongoDB.Default != "" {
				return errors.New("default value cannot be set in common, when it's set in database")
			}
			c.Results.MongoDB.Default = c.Results.Common.Default
			c.Results.FerretDB.Default = c.Results.Common.Default
		}

		if c.Results.Common.Stats != nil {
			if c.Results.FerretDB.Stats != nil || c.Results.MongoDB.Stats != nil {
				return errors.New("stats value cannot be set in common, when it's set in database")
			}
			c.Results.MongoDB.Stats = c.Results.Common.Stats
			c.Results.FerretDB.Stats = c.Results.Common.Stats
		}
	} else if c.Results.FerretDB == nil || c.Results.MongoDB == nil {
		return fmt.Errorf("both FerretDB and MongoDB results must be set (if common results are not set)")
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
