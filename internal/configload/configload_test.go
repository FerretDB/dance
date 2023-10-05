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
	"testing"

	"github.com/stretchr/testify/assert"

	ic "github.com/FerretDB/dance/internal/config"
)

func TestConvertBackend(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		b           *backend
		expected    *ic.TestConfig
		expectedErr error
	}{
		"DuplicatePrefix": {
			b:           &backend{Fail: []string{"a"}, Pass: []string{"a", "b"}},
			expected:    &ic.TestConfig{},
			expectedErr: errors.New("duplicate test name: \"a\""),
		},
		"InvalidStatus": {
			b:           &backend{Default: ic.Status("foo"), Fail: []string{"a"}},
			expected:    &ic.TestConfig{},
			expectedErr: errors.New("invalid status \"foo\""),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.b.convert(nil)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestIncludes(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		in          *backend
		includes    map[string][]string
		expected    *ic.TestConfig
		expectedErr error
	}{
		"IncludeFail": {
			in: &backend{
				Default:     ic.Status("pass"),
				Fail:        []string{"a"},
				IncludeFail: []string{"include_fail"},
			},
			includes: map[string][]string{
				"include_fail": {"x", "y", "z"},
			},
			expected: &ic.TestConfig{
				Fail: ic.Tests{
					Names: []string{"a", "x", "y", "z"},
				},
			},
			expectedErr: nil,
		},
		"IncludePass": {
			in: &backend{
				Default:     ic.Status("pass"),
				Pass:        []string{"x"},
				IncludePass: []string{"include_pass"},
			},
			includes: map[string][]string{
				"include_pass": {"a", "b", "c"},
			},
			expected: &ic.TestConfig{
				Pass: ic.Tests{
					Names: []string{"x", "a", "b", "c"},
				},
			},
			expectedErr: nil,
		},
		"IncludeIgnore": {
			in: &backend{
				Default:       ic.Status("ignore"),
				Ignore:        []string{"i"},
				IncludeIgnore: []string{"include_ignore"},
			},
			includes: map[string][]string{
				"include_ignore": {"a", "b", "c"},
			},
			expected: &ic.TestConfig{
				Ignore: ic.Tests{
					Names: []string{"i", "a", "b", "c"},
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

			assert.NoError(t, err)
		})
	}
}
