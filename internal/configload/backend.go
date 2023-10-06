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
	"sort"

	ic "github.com/FerretDB/dance/internal/config"
)

// backend represents the YAML-based configuration for database-specific test configurations.
//
//nolint:govet // we don't care about alignment there
type backend struct {
	Default ic.Status `yaml:"default"`
	Stats   stats     `yaml:"stats"`

	Fail   []string `yaml:"fail"`
	Skip   []string `yaml:"skip"`
	Pass   []string `yaml:"pass"`
	Ignore []string `yaml:"ignore"`

	IncludeFail   []string `yaml:"include_fail"`
	IncludeSkip   []string `yaml:"include_skip"`
	IncludePass   []string `yaml:"include_pass"`
	IncludeIgnore []string `yaml:"include_ignore"`
}

// convert converts backend configuration to [*config.TestConfig].
func (b *backend) convert(includes map[string][]string) (*ic.TestConfig, error) {
	if b == nil {
		panic("backend is nil")
	}

	t := &ic.TestConfig{
		Default: b.Default,
		Stats:   b.Stats.convert(),
		Fail:    ic.Tests{},
		Skip:    ic.Tests{},
		Pass:    ic.Tests{},
		Ignore:  ic.Tests{},
	}

	if t.Default == "" {
		t.Default = ic.Pass
	}

	expected := []ic.Status{ic.Fail, ic.Skip, ic.Pass, ic.Ignore} // no unknown
	if !slices.Contains(expected, t.Default) {
		return nil, fmt.Errorf("invalid status %q", t.Default)
	}

	names := make(map[string]struct{})

	for dst, src := range map[*[]string][]string{
		&t.Fail.Names:   b.Fail,
		&t.Skip.Names:   b.Skip,
		&t.Pass.Names:   b.Pass,
		&t.Ignore.Names: b.Ignore,
	} {
		for _, name := range src {
			if _, ok := names[name]; ok {
				return nil, fmt.Errorf("duplicate test name: %q", name)
			}
			names[name] = struct{}{}
		}

		*dst = append(*dst, src...)
	}

	for dst, src := range map[*[]string][]string{
		&t.Fail.Names:   b.IncludeFail,
		&t.Skip.Names:   b.IncludeSkip,
		&t.Pass.Names:   b.IncludePass,
		&t.Ignore.Names: b.IncludeIgnore,
	} {
		for _, group := range src {
			for _, name := range includes[group] {
				if _, ok := names[name]; ok {
					return nil, fmt.Errorf("duplicate test name: %q", name)
				}
				names[name] = struct{}{}

				*dst = append(*dst, name)
			}
		}

		sort.Strings(*dst)
	}

	return t, nil
}
