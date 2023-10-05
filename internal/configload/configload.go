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
		// Includes is a mapping that allows us to merge sequences together,
		// which is currently not possible in the YAML spec - https://github.com/yaml/yaml/issues/48
		Includes   map[string][]string `yaml:"includes"`
		PostgreSQL *backend            `yaml:"postgresql"`
		SQLite     *backend            `yaml:"sqlite"`
		MongoDB    *backend            `yaml:"mongodb"`
	} `yaml:"results"`
}

// Load reads and validates the configuration from a YAML file,
// returning a pointer to the internal configuration struct *config.Config.
// Any error encountered during the process is also returned.
func Load(file string) (*ic.Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	var cfg config
	if err = d.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	postgreSQL, err := cfg.Results.PostgreSQL.convert(cfg.Results.Includes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PostgreSQL config: %w", err)
	}

	sqLite, err := cfg.Results.SQLite.convert(cfg.Results.Includes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert SQLite config: %w", err)
	}

	mongoDB, err := cfg.Results.MongoDB.convert(cfg.Results.Includes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert MongoDB config: %w", err)
	}

	return &ic.Config{
		Runner: cfg.Runner,
		Dir:    cfg.Dir,
		Args:   cfg.Args,
		Results: ic.Results{
			PostgreSQL: postgreSQL,
			SQLite:     sqLite,
			MongoDB:    mongoDB,
		},
	}, nil
}
