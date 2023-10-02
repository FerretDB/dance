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

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"

	ic "github.com/FerretDB/dance/internal/config"
)

func TestFillAndValidate(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		in          *config
		expectedErr error
	}{
		"Nil": {
			in:          &config{},
			expectedErr: nil,
		},
		"InvalidDefault": {
			in: &config{
				Results: struct {
					Includes   map[string][]string "yaml:\"includes\""
					PostgreSQL *testConfig         "yaml:\"ferretdb\""
					SQLite     *testConfig         "yaml:\"sqlite\""
					MongoDB    *testConfig         "yaml:\"mongodb\""
				}{
					Includes:   map[string][]string{},
					PostgreSQL: &testConfig{Default: (*ic.Status)(pointer.ToString("foo"))},
				},
			},
			expectedErr: fmt.Errorf("invalid default result: %q", "foo"),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.in.fillAndValidate()
			if err != nil {
				assert.Equal(t, err, tc.expectedErr)
			}
		})
	}
}

func TestConvertAndValidate(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		in          *config
		expectedErr error
	}{
		"DuplicatePrefix": {
			in: &config{
				Results: struct {
					Includes   map[string][]string "yaml:\"includes\""
					PostgreSQL *testConfig         "yaml:\"ferretdb\""
					SQLite     *testConfig         "yaml:\"sqlite\""
					MongoDB    *testConfig         "yaml:\"mongodb\""
				}{
					Includes: map[string][]string{},
					PostgreSQL: &testConfig{
						Default: (*ic.Status)(pointer.ToString("fail")),
						Fail:    []string{"a"},
						Pass:    []string{"a"}},
				},
			},
			expectedErr: fmt.Errorf("duplicate test or prefix: %q", "a"),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.in.convertAndValidate()
			if err != nil {
				assert.Equal(t, err, tc.expectedErr)
			}
		})
	}
}

func TestIncludes(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		in          *testConfig
		includes    map[string][]string
		expected    *ic.TestConfig
		expectedErr error
	}{
		"IncludeFail": {
			in: &testConfig{
				Default:     (*ic.Status)(pointer.ToString("fail")),
				Fail:        []string{"a"},
				IncludeFail: []string{"include_fail"},
			},
			includes: map[string][]string{
				"include_fail": {"x", "y", "z"},
			},
			expected: &ic.TestConfig{
				Fail: ic.Tests{
					Names: []string{"x", "y", "z", "a"},
				},
			},
			expectedErr: nil,
		},
		"IncludePass": {
			in: &testConfig{
				Default:     (*ic.Status)(pointer.ToString("pass")),
				Pass:        []string{"x"},
				IncludePass: []string{"include_pass"},
			},
			includes: map[string][]string{
				"include_pass": {"a", "b", "c"},
			},
			expected: &ic.TestConfig{
				Pass: ic.Tests{
					Names: []string{"a", "b", "c", "x"},
				},
			},
			expectedErr: nil,
		},
		"IncludeIgnore": {
			in: &testConfig{
				Default:       (*ic.Status)(pointer.ToString("ignore")),
				Ignore:        []string{"i"},
				IncludeIgnore: []string{"include_ignore"},
			},
			includes: map[string][]string{
				"include_ignore": {"a", "b", "c"},
			},
			expected: &ic.TestConfig{
				Ignore: ic.Tests{
					Names: []string{"a", "b", "c", "i"},
				},
			},
			expectedErr: nil,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out, err := tc.in.convert(tc.includes)

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				return
			}

			if out.Fail.Names != nil {
				assert.Equal(t, tc.expected.Fail, out.Fail)
			}

			if out.Pass.Names != nil {
				assert.Equal(t, tc.expected.Pass, out.Pass)
			}

			if out.Ignore.Names != nil {
				assert.Equal(t, tc.expected.Ignore, out.Ignore)
			}
		})
	}
}
