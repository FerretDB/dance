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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	ic "github.com/FerretDB/dance/internal/config"
)

func TestMergeCommon(t *testing.T) {
	t.Parallel()

	//nolint:govet // we don't care about alignment there
	for name, tc := range map[string]struct {
		common      *ic.TestConfig
		config1     *ic.TestConfig
		config2     *ic.TestConfig
		expectedErr error
	}{
		"AllNil": {
			common:      nil,
			config1:     nil,
			config2:     nil,
			expectedErr: fmt.Errorf("all database-specific results must be set (if common results are not set)"),
		},
		"AllPass": {
			common: &ic.TestConfig{
				Pass: ic.Tests{Names: []string{"a"}},
			},
			config1: &ic.TestConfig{
				Pass: ic.Tests{Names: []string{"e"}},
			},
			config2: &ic.TestConfig{
				Pass: ic.Tests{Names: []string{"i"}},
			},
			expectedErr: nil,
		},
		"DuplicatesPass": {
			common: &ic.TestConfig{
				Pass: ic.Tests{Names: []string{"a"}},
			},
			config1: &ic.TestConfig{
				Pass: ic.Tests{Names: []string{"e", "a"}},
			},
			config2: &ic.TestConfig{
				Pass: ic.Tests{Names: []string{"i"}},
			},
			expectedErr: fmt.Errorf("duplicate test or prefix: \"a\""),
		},
		"FailOutputRegexPattern": {
			common: &ic.TestConfig{
				Fail: ic.Tests{OutputRegexPattern: []string{"skipping due to environmental constraints.*"}},
			},
			config1: &ic.TestConfig{
				Fail: ic.Tests{OutputRegexPattern: []string{"a"}},
			},
			config2: &ic.TestConfig{
				Fail: ic.Tests{OutputRegexPattern: []string{"b"}},
			},
			expectedErr: nil,
		},
		"AllStats": {
			common: &ic.TestConfig{
				Stats: &ic.Stats{},
			},
			config1: &ic.TestConfig{
				Stats: &ic.Stats{},
			},
			config2: &ic.TestConfig{
				Stats: &ic.Stats{},
			},
			expectedErr: errors.New("stats value cannot be set in common, when it's set in database"),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := mergeCommon(tc.common, tc.config1, tc.config2)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				return
			}

			assert.NoError(t, err)

			for _, tests := range []struct {
				actual ic.Tests
			}{
				{tc.config1.Pass},
				{tc.config2.Pass},
			} {
				for _, name := range tc.common.Pass.Names {
					assert.Contains(t, tests.actual.Names, name)
				}
				for _, name := range tc.common.Pass.NameRegexPattern {
					assert.Contains(t, tests.actual.NameRegexPattern, name)
				}
				for _, name := range tc.common.Pass.NameNotRegexPattern {
					assert.Contains(t, tests.actual.NameNotRegexPattern, name)
				}
				for _, name := range tc.common.Pass.OutputRegexPattern {
					assert.Contains(t, tests.actual.OutputRegexPattern, name)
				}
			}

			for _, tests := range []struct {
				actual ic.Tests
			}{
				{tc.config1.Fail},
				{tc.config2.Fail},
			} {
				for _, name := range tc.common.Fail.Names {
					assert.Contains(t, tests.actual.Names, name)
				}
				for _, name := range tc.common.Fail.NameRegexPattern {
					assert.Contains(t, tests.actual.NameRegexPattern, name)
				}
				for _, name := range tc.common.Fail.NameNotRegexPattern {
					assert.Contains(t, tests.actual.NameNotRegexPattern, name)
				}
				for _, name := range tc.common.Fail.OutputRegexPattern {
					assert.Contains(t, tests.actual.OutputRegexPattern, name)
				}
			}
		})
	}
}
