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
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Runner string      `yaml:"runner"`
	Args   []string    `yaml:"args"`
	Tests  TestsConfig `yaml:"tests"`
}

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

	if err = c.validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) validate() error {
	return nil
}
