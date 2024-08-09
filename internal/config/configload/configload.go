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
	"log"
	"os"

	"github.com/go-viper/mapstructure/v2"
	_ "github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal/config"
)

// file is used to unmarshal configuration from YAML.
//
//nolint:vet // for readability
type file struct {
	Runner  config.RunnerType `yaml:"runner"`
	Config  map[string]any    `yaml:"config"`
	Results struct {
		PostgreSQL *result `yaml:"postgresql"`
		SQLite     *result `yaml:"sqlite"`
		MongoDB    *result `yaml:"mongodb"`
	} `yaml:"results"`
}

// Load reads and validates the configuration from a YAML file.
func Load(name string) (*config.Config, error) {
	f2, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f2.Close()

	var f file
	yd := yaml.NewDecoder(f2)
	yd.KnownFields(true)

	if err = yd.Decode(&f); err != nil {
		return nil, fmt.Errorf("failed to decode YAML config: %w", err)
	}

	var cfg runnerConfig
	switch f.Runner {
	case config.RunnerTypeCommand:
		cfg = &runnerConfigCommand{}
	case config.RunnerTypeGoTest:
		fallthrough
		// cfg = &runnerConfigGoTest{}
	case config.RunnerTypeJSTest:
		fallthrough
		// cfg = &runnerConfigJSTest{}
	case config.RunnerTypeYCSB:
		fallthrough
		// cfg = &runnerConfigYCSB{}
	default:
		err = fmt.Errorf("unknown runner type %q", f.Runner)
	}
	if err != nil {
		return nil, err
	}

	md, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		ErrorUnset:  true,
		Result:      cfg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decode runner config: %w", err)
	}

	log.Printf("%+v", f.Config)

	if err = md.Decode(f.Config); err != nil {
		return nil, fmt.Errorf("failed to decode runner config 2: %w", err)
	}

	log.Printf("%+v", cfg)

	postgreSQL, err := f.Results.PostgreSQL.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert PostgreSQL config: %w", err)
	}

	sqLite, err := f.Results.SQLite.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert SQLite config: %w", err)
	}

	mongoDB, err := f.Results.MongoDB.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert MongoDB config: %w", err)
	}

	return &config.Config{
		Runner: f.Runner,
		Dir:    cfg.GetDir(),
		Args:   cfg.GetArgs(),
		Results: config.Results{
			PostgreSQL: postgreSQL,
			SQLite:     sqLite,
			MongoDB:    mongoDB,
		},
	}, nil
}
