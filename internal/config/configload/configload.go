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

// project is used to unmarshal configuration from YAML.
//
//nolint:vet // for readability
type project struct {
	Runner  config.RunnerType `yaml:"runner"`
	Config  yaml.Node         `yaml:"config"`
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
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var p project
	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	if err = d.Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode YAML config: %w", err)
	}

	var cfg runnerConfig
	switch p.Runner {
	case config.RunnerTypeCommand:
		cfg = &runnerConfigCommand{}
	case config.RunnerTypeGoTest:
		cfg = &runnerConfigGoTest{}
	case config.RunnerTypeJSTest:
		cfg = &runnerConfigJSTest{}
	case config.RunnerTypeYCSB:
		cfg = &runnerConfigYCSB{}
	default:
		err = fmt.Errorf("unknown runner type %q", p.Runner)
	}
	if err != nil {
		return nil, err
	}

	if err = p.Config.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode runner config: %w", err)
	}

	postgreSQL, err := p.Results.PostgreSQL.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert PostgreSQL config: %w", err)
	}

	sqLite, err := p.Results.SQLite.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert SQLite config: %w", err)
	}

	mongoDB, err := p.Results.MongoDB.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert MongoDB config: %w", err)
	}

	return &config.Config{
		Runner: p.Runner,
		Dir:    cfg.GetDir(),
		Args:   cfg.GetArgs(),
		Results: config.Results{
			PostgreSQL: postgreSQL,
			SQLite:     sqLite,
			MongoDB:    mongoDB,
		},
	}, nil
}
