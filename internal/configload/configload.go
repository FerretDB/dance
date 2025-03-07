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
	"net"
	"net/url"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal/config"
)

// DBs contains MongoDB URIs for different databases.
var DBs = map[string]string{
	"mongodb":              "mongodb://127.0.0.1:37001/",
	"mongodb-secured":      "mongodb://username:password@127.0.0.1:37002/?authSource=admin",
	"ferretdb":             "mongodb://127.0.0.1:27001/",
	"ferretdb-secured":     "mongodb://username:password@127.0.0.1:27002/?authSource=admin",
	"ferretdb-dev":         "mongodb://127.0.0.1:27003/",
	"ferretdb-dev-secured": "mongodb://username:password@127.0.0.1:27004/?authSource=admin",
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
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config file: %w", err)
	}

	return loadContent(string(b), db)
}

// templateData returns a map with template data for the given MongoDB URI.
func templateData(uri url.URL) (map[string]any, error) {
	anonymousURI := uri
	anonymousURI.User = nil

	plainURI := uri
	q := plainURI.Query()
	q.Set("authMechanism", "PLAIN")

	if plainURI.User == nil {
		plainURI.User = url.UserPassword("dummy", "dummy")

		q.Set("authSource", "admin")
	}

	plainURI.RawQuery = q.Encode()

	sha1URI := uri
	q = sha1URI.Query()
	q.Set("authMechanism", "SCRAM-SHA-1")

	if sha1URI.User == nil {
		sha1URI.User = url.UserPassword("dummy", "dummy")

		q.Set("authSource", "admin")
	}

	sha1URI.RawQuery = q.Encode()

	sha256URI := uri
	q = sha256URI.Query()
	q.Set("authMechanism", "SCRAM-SHA-256")

	if sha256URI.User == nil {
		sha256URI.User = url.UserPassword("dummy", "dummy")

		q.Set("authSource", "admin")
	}

	sha256URI.RawQuery = q.Encode()

	dockerHostURI := uri

	_, port, err := net.SplitHostPort(dockerHostURI.Host)
	if err != nil {
		return nil, err
	}

	dockerHostURI.Host = net.JoinHostPort("host.docker.internal", port)

	return map[string]any{
		"MONGODB_URI":             uri.String(),
		"MONGODB_URI_ANONYMOUS":   anonymousURI.String(),
		"MONGODB_URI_PLAIN":       plainURI.String(),
		"MONGODB_URI_SHA1":        sha1URI.String(),
		"MONGODB_URI_SHA256":      sha256URI.String(),
		"MONGODB_URI_DOCKER_HOST": dockerHostURI.String(),
	}, nil
}

// loadContent reads and validates project configuration for the given database from the YAML content.
func loadContent(content, db string) (*config.Config, error) {
	mongodbURI, ok := DBs[db]
	if !ok {
		return nil, fmt.Errorf("unknown database %q", db)
	}

	if mongodbURI == "" {
		return nil, fmt.Errorf("no MongoDB URI for %q", db)
	}

	u, err := url.Parse(mongodbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MongoDB URI %q for %q: %w", mongodbURI, db, err)
	}

	t, err := template.New("").Option("missingkey=error").Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project config file template: %w", err)
	}

	data, err := templateData(*u)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
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
		p = &runnerParamsGoTest{}
	case config.RunnerTypeYCSB:
		p = &runnerParamsYCSB{}
	default:
		err = fmt.Errorf("unknown runner type %q", pc.Runner)
	}

	if err != nil {
		return nil, err
	}

	if err = pc.Params.Decode(p); err != nil {
		return nil, fmt.Errorf("failed to decode runner parameters: %w", err)
	}

	params, err := p.convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert runner parameters: %w", err)
	}

	res := pc.Results[db]
	if res == nil {
		return nil, nil
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
