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

package configload

import (
	"fmt"
	"slices"

	"github.com/FerretDB/dance/internal/config"
)

// result represents expected results for specific database in the project configuration YAML file.
//
//nolint:vet // for readability
type result struct {
	Default config.Status `yaml:"default"` // defaults to pass
	Stats   stats         `yaml:"stats"`

	// test names
	Fail   []string `yaml:"fail"`
	Skip   []string `yaml:"skip"`
	Pass   []string `yaml:"pass"`
	Ignore []string `yaml:"ignore"`
}

// convert converts result to [*config.TestConfig].
func (r *result) convert() (*config.TestConfig, error) {
	if r == nil {
		panic("backend is nil")
	}

	t := &config.TestConfig{
		Default: r.Default,
		Stats:   r.Stats.convert(),
	}

	if t.Default == "" {
		t.Default = config.Pass
	}

	expected := []config.Status{config.Fail, config.Skip, config.Pass, config.Ignore} // no unknown
	if !slices.Contains(expected, t.Default) {
		return nil, fmt.Errorf("invalid default status %q", t.Default)
	}

	names := make(map[string]struct{})

	for dst, src := range map[*[]string][]string{
		&t.Fail:   r.Fail,
		&t.Skip:   r.Skip,
		&t.Pass:   r.Pass,
		&t.Ignore: r.Ignore,
	} {
		for _, name := range src {
			if _, ok := names[name]; ok {
				return nil, fmt.Errorf("duplicate test name: %q", name)
			}
			names[name] = struct{}{}
		}

		*dst = append(*dst, src...)
	}

	return t, nil
}
