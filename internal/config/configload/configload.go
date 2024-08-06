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

	"github.com/FerretDB/dance/internal/config"
)

// file is used to unmarshal configuration from YAML.
//
//nolint:vet // for readability
type file struct {
	Runner  config.RunnerType `yaml:"runner"`
	Dir     string            `yaml:"dir"`
	Args    []string          `yaml:"args"`
	Results struct {
		PostgreSQL *result `yaml:"postgresql"`
		SQLite     *result `yaml:"sqlite"`
		MongoDB    *result `yaml:"mongodb"`
	} `yaml:"results"`
}

// Load reads and validates the configuration from a YAML file.
func Load(name string) (*config.Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	var cfg file
	if err = d.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	postgreSQL, err := cfg.Results.PostgreSQL.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert PostgreSQL config: %w", err)
	}

	sqLite, err := cfg.Results.SQLite.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert SQLite config: %w", err)
	}

	mongoDB, err := cfg.Results.MongoDB.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert MongoDB config: %w", err)
	}

	return &config.Config{
		Runner: cfg.Runner,
		Dir:    cfg.Dir,
		Args:   cfg.Args,
		Results: config.Results{
			PostgreSQL: postgreSQL,
			SQLite:     sqLite,
			MongoDB:    mongoDB,
		},
	}, nil
}
