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

package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzNextPrefix(f *testing.F) {
	type testCase struct {
		paths []string
	}

	for _, tc := range []testCase{{
		paths: []string{
			"topology/TestCMAPSpec/pool-checkin-destroy-closed.json",
			"topology/TestCMAPSpec/pool-checkin-destroy-closed.",
			"topology/TestCMAPSpec/pool-checkin-destroy-closed",
			"topology/TestCMAPSpec/",
			"topology/TestCMAPSpec",
			"topology/",
			"topology",
			"",
		},
	}} {
		for i, path := range tc.paths[:len(tc.paths)-1] {
			expected := tc.paths[i+1]
			actual := nextPrefix(path)
			assert.Equal(f, expected, actual, "path = %q", path)

			f.Add(path)
			f.Add(expected)
			f.Add(actual)
		}
	}

	f.Fuzz(func(t *testing.T, path string) {
		for path != "" {
			next := nextPrefix(path)
			require.NotEqual(t, next, path)
			path = next
		}
	})
}

func TestMergeTestConfigs(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		in          *Results
		expected    *Results
		expectedErr error
	}{
		"MergeTestConfigsFilled": {
			in: &Results{
				Common: &TestConfig{
					Pass:   Tests{Names: []string{"a", "b"}},
					Skip:   Tests{Names: []string{"c", "d"}},
					Fail:   Tests{Names: []string{"e", "f"}},
					Ignore: Tests{Names: []string{"g", "h"}},
				},
				FerretDB: &TestConfig{
					Pass: Tests{Names: []string{"1", "2"}},
					Skip: Tests{Names: []string{"3", "4"}},
					Fail: Tests{Names: []string{"5"}},
				},
				MongoDB: &TestConfig{
					Pass:   Tests{Names: []string{"A", "B"}},
					Skip:   Tests{Names: []string{"C"}},
					Fail:   Tests{Names: []string{"D", "E"}},
					Ignore: Tests{Names: []string{"x", "z"}},
				},
			},
			expected: &Results{
				Common: &TestConfig{
					Pass:   Tests{Names: []string{"a", "b"}},
					Skip:   Tests{Names: []string{"c", "d"}},
					Fail:   Tests{Names: []string{"e", "f"}},
					Ignore: Tests{Names: []string{"g", "h"}},
				},
				FerretDB: &TestConfig{
					Pass:   Tests{Names: []string{"1", "2", "a", "b"}},
					Skip:   Tests{Names: []string{"3", "4", "c", "d"}},
					Fail:   Tests{Names: []string{"5", "e", "f"}},
					Ignore: Tests{Names: []string{"g", "h"}},
				},
				MongoDB: &TestConfig{
					Pass:   Tests{Names: []string{"A", "B", "a", "b"}},
					Skip:   Tests{Names: []string{"C", "c", "d"}},
					Fail:   Tests{Names: []string{"D", "E", "e", "f"}},
					Ignore: Tests{Names: []string{"g", "h", "x", "z"}},
				},
			},
		},
		"MergeTestConfigsNotSet": {
			in: &Results{
				Common:   nil,
				FerretDB: &TestConfig{},
			},
			expectedErr: errors.New("all database-specific results must be set (if common results are not set)"),
		},
		"MergeTestConfigsDuplicatesPass": {
			in: &Results{
				Common: &TestConfig{
					Pass: Tests{Names: []string{"a"}},
				},
				FerretDB: &TestConfig{
					Pass: Tests{Names: []string{"a", "b"}},
				},
				MongoDB: &TestConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"MergeTestConfigsDuplicatesSkip": {
			in: &Results{
				Common: &TestConfig{
					Skip: Tests{Names: []string{"a"}},
				},
				FerretDB: &TestConfig{
					Skip: Tests{Names: []string{"a", "b"}},
				},
				MongoDB: &TestConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"MergeTestConfigsDuplicatesFail": {
			in: &Results{
				Common: &TestConfig{
					Fail: Tests{Names: []string{"a"}},
				},
				FerretDB: &TestConfig{
					Fail: Tests{Names: []string{"a", "b"}},
				},
				MongoDB: &TestConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"MergeTestConfigsDuplicatesAll": {
			in: &Results{
				FerretDB: &TestConfig{
					Pass:   Tests{Names: []string{"x"}},
					Skip:   Tests{Names: []string{"x"}},
					Fail:   Tests{Names: []string{"x"}},
					Ignore: Tests{Names: []string{"x"}},
				},
				MongoDB: &TestConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"x\""),
		},
		"MergeTestConfigsDefault": {
			in: &Results{
				Common: &TestConfig{
					Default: "pass",
				},
				FerretDB: &TestConfig{},
				MongoDB:  &TestConfig{},
			},
			expected: &Results{
				Common: &TestConfig{
					Default: "pass",
				},
				FerretDB: &TestConfig{
					Default: "pass",
				},
				MongoDB: &TestConfig{
					Default: "pass",
				},
			},
		},
		// XXX error is not returned here
		"MergeTestConfigsDefaultDuplicate": {
			in: &Results{
				Common: &TestConfig{
					Default: "pass",
				},
				FerretDB: &TestConfig{Default: "fail"},
				MongoDB:  &TestConfig{},
			},
			expectedErr: errors.New("default value cannot be set in common, when it's set in database"),
		},
		"MergeTestConfigsStats": {
			in: &Results{
				Common: &TestConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
				FerretDB: &TestConfig{},
				MongoDB:  &TestConfig{},
			},
			expected: &Results{
				Common: &TestConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
				FerretDB: &TestConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
				MongoDB: &TestConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
			},
		},
		"MergeTestConfigsStatsDuplicate": {
			in: &Results{
				Common: &TestConfig{
					Stats: &Stats{},
				},
				FerretDB: &TestConfig{
					Stats: &Stats{},
				},
				MongoDB: &TestConfig{
					Stats: &Stats{},
				},
			},
			expectedErr: errors.New("stats value cannot be set in common, when it's set in database"),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var c Config
			c.Results = *tc.in

			err := MergeTestConfigs(c.Results.Common, c.Results.FerretDB, c.Results.MongoDB)

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				return
			}

			assert.NoError(t, err)

			for _, tests := range []struct {
				expected Tests
				actual   Tests
			}{
				{tc.expected.FerretDB.Pass, tc.in.FerretDB.Pass},
				{tc.expected.FerretDB.Skip, tc.in.FerretDB.Skip},
				{tc.expected.FerretDB.Fail, tc.in.FerretDB.Fail},
				{tc.expected.FerretDB.Ignore, tc.in.FerretDB.Ignore},

				{tc.expected.MongoDB.Pass, tc.in.MongoDB.Pass},
				{tc.expected.MongoDB.Skip, tc.in.MongoDB.Skip},
				{tc.expected.MongoDB.Fail, tc.in.MongoDB.Fail},
				{tc.expected.MongoDB.Ignore, tc.in.MongoDB.Ignore},
			} {
				for _, item := range tests.expected.Names {
					assert.Contains(t, tests.actual.Names, item)
				}
				assert.Equal(t, len(tests.expected.Names), len(tests.actual.Names))

				for _, item := range tests.expected.NameRegexPattern {
					assert.Contains(t, tests.actual.NameRegexPattern, item)
				}
				assert.Equal(t, len(tests.expected.NameRegexPattern), len(tests.actual.NameRegexPattern))

				for _, item := range tests.expected.NameNotRegexPattern {
					assert.Contains(t, tests.actual.NameNotRegexPattern, item)
				}
				assert.Equal(t, len(tests.expected.NameNotRegexPattern), len(tests.actual.NameNotRegexPattern))

				for _, item := range tests.expected.OutputRegexPattern {
					assert.Contains(t, tests.actual.OutputRegexPattern, item)
				}
				assert.Equal(t, len(tests.expected.OutputRegexPattern), len(tests.actual.OutputRegexPattern))
			}
		})
	}
}
