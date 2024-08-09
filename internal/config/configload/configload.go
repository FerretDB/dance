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

// Package configload provides functionality for loading and validating project configuration from YAML files.
package configload

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal/config"
)

// projectConfig represents YAML project configuration file.
//
//nolint:vet // for readability
type projectConfig struct {
	Runner  config.RunnerType `yaml:"runner"`
	Params  yaml.Node         `yaml:"params"`
	Results struct {
		PostgreSQL *result `yaml:"postgresql"`
		SQLite     *result `yaml:"sqlite"`
		MongoDB    *result `yaml:"mongodb"`
	} `yaml:"results"`
}

// Load reads and validates project configuration from a YAML file.
func Load(name string) (*config.Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open project config file: %w", err)
	}
	defer f.Close()

	var pc projectConfig
	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	if err = d.Decode(&pc); err != nil {
		return nil, fmt.Errorf("failed to decode project config: %w", err)
	}

	var p runnerParams
	switch pc.Runner {
	case config.RunnerTypeCommand:
		p = &runnerParamsCommand{}
	case config.RunnerTypeGoTest:
		fallthrough
	case config.RunnerTypeJSTest:
		fallthrough
	case config.RunnerTypeYCSB:
		fallthrough
	default:
		err = fmt.Errorf("unknown runner type %q", pc.Runner)
	}
	if err != nil {
		return nil, err
	}

	if err = pc.Params.Decode(p); err != nil {
		return nil, fmt.Errorf("failed to decode runner parameters: %w", err)
	}

	postgreSQL, err := pc.Results.PostgreSQL.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert PostgreSQL results: %w", err)
	}

	sqLite, err := pc.Results.SQLite.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert SQLite results: %w", err)
	}

	mongoDB, err := pc.Results.MongoDB.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert MongoDB results: %w", err)
	}

	return &config.Config{
		Runner: pc.Runner,
		Dir:    p.GetDir(),
		Args:   p.GetArgs(),
		Results: config.Results{
			PostgreSQL: postgreSQL,
			SQLite:     sqLite,
			MongoDB:    mongoDB,
		},
	}, nil
}
