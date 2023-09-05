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
		expected    []*ic.TestConfig
		expectedErr error
	}{
		"AllNil": {
			common:      nil,
			config1:     nil,
			config2:     nil,
			expected:    nil,
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
			expected: []*ic.TestConfig{
				{
					Pass: ic.Tests{Names: []string{"e", "a"}},
				},
				{
					Pass: ic.Tests{Names: []string{"i", "a"}},
				},
			},
			expectedErr: nil,
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
				expected []*ic.TestConfig
				actual   ic.Tests
			}{
				{tc.expected, tc.common.Pass},
			} {
				for _, item := range tests.expected {
					for _, name := range tc.common.Pass.Names {
						assert.Contains(t, item.Pass.Names, name)
					}
				}
			}
		})
	}
}
