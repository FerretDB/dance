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
	"bytes"
	"fmt"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal/config"
)

// uris contains MongoDB URIs for different databases.
var uris = map[string]string{
	"mongodb":             "mongodb://127.0.0.1:47017/",
	"mongodb-secured":     "mongodb://127.0.0.1:47018/",
	"ferretdb-postgresql": "mongodb://127.0.0.1:27001/",
	"ferretdb-sqlite":     "mongodb://127.0.0.1:27002/",
}

// projectConfig represents project configuration YAML file.
//
//nolint:vet // for readability
type projectConfig struct {
	Runner  config.RunnerType           `yaml:"runner"`
	Params  yaml.Node                   `yaml:"params"`
	Results map[string]*expectedResults `yaml:"results"`
}

// Load reads and validates project configuration for the given database from the YAML file.
func Load(file, db string) (*config.Config, error) {
	uri, ok := uris[db]
	if !ok {
		return nil, fmt.Errorf("unknown database %q", db)
	}
	if uri == "" {
		return nil, fmt.Errorf("no MongoDB URI for %q", db)
	}

	t, err := template.ParseFiles(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project config file template: %w", err)
	}

	var buf bytes.Buffer
	data := map[string]any{
		"MONGODB_URI": uri,
	}

	if err = t.Option("missingkey=error").Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute project config file template: %w", err)
	}

	var pc projectConfig
	d := yaml.NewDecoder(&buf)
	d.KnownFields(true)

	if err = d.Decode(&pc); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
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

	params := p.convert()

	res := pc.Results[db]
	if res == nil {
		return nil, fmt.Errorf("no results configuration for %q", db)
	}

	results, err := res.convert()
	if err != nil {
		return nil, fmt.Errorf("invalid results configuration for %q: %w", db, err)
	}

	return &config.Config{
		Runner:  pc.Runner,
		Params:  params,
		Results: results,
	}, nil
}
