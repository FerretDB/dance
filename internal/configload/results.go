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

// expectedResults represents expected results for specific database in the project configuration YAML file.
//
//nolint:vet // for readability
type expectedResults struct {
	Default config.Status `yaml:"default"` // defaults to pass
	Stats   stats         `yaml:"stats"`

	// test names
	Fail   []string `yaml:"fail"`
	Skip   []string `yaml:"skip"`
	Pass   []string `yaml:"pass"`
	Ignore []string `yaml:"ignore"`
}

// convert converts result to [*config.ExpectedResults].
func (r *expectedResults) convert() (*config.ExpectedResults, error) {
	if r == nil {
		panic("result is nil")
	}

	res := &config.ExpectedResults{
		Default: r.Default,
		Stats:   r.Stats.convert(),
	}

	if res.Default == "" {
		res.Default = config.Pass
	}

	statuses := []config.Status{config.Fail, config.Skip, config.Pass, config.Ignore} // no unknown
	if !slices.Contains(statuses, res.Default) {
		return nil, fmt.Errorf("invalid default status %q", res.Default)
	}

	names := make(map[string]struct{})

	for dst, src := range map[*[]string][]string{
		&res.Fail:   r.Fail,
		&res.Skip:   r.Skip,
		&res.Pass:   r.Pass,
		&res.Ignore: r.Ignore,
	} {
		for _, name := range src {
			if _, ok := names[name]; ok {
				return nil, fmt.Errorf("duplicate test name: %q", name)
			}
			names[name] = struct{}{}
		}

		*dst = append(*dst, src...)
	}

	return res, nil
}
