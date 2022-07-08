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

package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillAndValidate(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		in          *Results
		expected    *Results
		expectedErr error
	}{
		"FillAndValidateFilled": {
			in: &Results{
				Common: &TestsConfig{
					Pass: Tests{TestNames: []string{"a", "b"}},
					Skip: Tests{TestNames: []string{"c", "d"}},
					Fail: Tests{TestNames: []string{"e", "f"}},
				},
				FerretDB: &TestsConfig{
					Pass: Tests{TestNames: []string{"1", "2"}},
					Skip: Tests{TestNames: []string{"3", "4"}},
					Fail: Tests{TestNames: []string{"5"}},
				},
				MongoDB: &TestsConfig{
					Pass: Tests{TestNames: []string{"A", "B"}},
					Skip: Tests{TestNames: []string{"C"}},
					Fail: Tests{TestNames: []string{"D", "E"}},
				},
			},
			expected: &Results{
				Common: &TestsConfig{
					Pass: Tests{TestNames: []string{"a", "b"}},
					Skip: Tests{TestNames: []string{"c", "d"}},
					Fail: Tests{TestNames: []string{"e", "f"}},
				},
				FerretDB: &TestsConfig{
					Pass: Tests{TestNames: []string{"1", "2", "a", "b"}},
					Skip: Tests{TestNames: []string{"3", "4", "c", "d"}},
					Fail: Tests{TestNames: []string{"5", "e", "f"}},
				},
				MongoDB: &TestsConfig{
					Pass: Tests{TestNames: []string{"A", "B", "a", "b"}},
					Skip: Tests{TestNames: []string{"C", "c", "d"}},
					Fail: Tests{TestNames: []string{"D", "E", "e", "f"}},
				},
			},
		},
		"FillAndValidateNotSet": {
			in: &Results{
				Common:   nil,
				FerretDB: &TestsConfig{},
			},
			expectedErr: errors.New("both FerretDB and MongoDB results must be set (if common results are not set)"),
		},
		"FillAndValidateDuplicatesPass": {
			in: &Results{
				Common: &TestsConfig{
					Pass: Tests{TestNames: []string{"a"}},
				},
				FerretDB: &TestsConfig{
					Pass: Tests{TestNames: []string{"a", "b"}},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"FillAndValidateDuplicatesSkip": {
			in: &Results{
				Common: &TestsConfig{
					Skip: Tests{TestNames: []string{"a"}},
				},
				FerretDB: &TestsConfig{
					Skip: Tests{TestNames: []string{"a", "b"}},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"FillAndValidateDuplicatesFail": {
			in: &Results{
				Common: &TestsConfig{
					Fail: Tests{TestNames: []string{"a"}},
				},
				FerretDB: &TestsConfig{
					Fail: Tests{TestNames: []string{"a", "b"}},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"FillAndValidateDuplicatesAll": {
			in: &Results{
				FerretDB: &TestsConfig{
					Pass: Tests{TestNames: []string{"a"}},
					Skip: Tests{TestNames: []string{"a"}},
					Fail: Tests{TestNames: []string{"a"}},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"FillAndValidateDefault": {
			in: &Results{
				Common: &TestsConfig{
					Default: "pass",
				},
				FerretDB: &TestsConfig{},
				MongoDB:  &TestsConfig{},
			},
			expected: &Results{
				Common: &TestsConfig{
					Default: "pass",
				},
				FerretDB: &TestsConfig{
					Default: "pass",
				},
				MongoDB: &TestsConfig{
					Default: "pass",
				},
			},
		},
		"FillAndValidateDefaultDuplicate": {
			in: &Results{
				Common: &TestsConfig{
					Default: "pass",
				},
				FerretDB: &TestsConfig{Default: "fail"},
				MongoDB:  &TestsConfig{},
			},
			expectedErr: errors.New("default value cannot be set in common, when it's set in database"),
		},
		"FillAndValidateStats": {
			in: &Results{
				Common: &TestsConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
				FerretDB: &TestsConfig{},
				MongoDB:  &TestsConfig{},
			},
			expected: &Results{
				Common: &TestsConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
				FerretDB: &TestsConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
				MongoDB: &TestsConfig{
					Stats: &Stats{1, 2, 3, 4, 5, 6, 7},
				},
			},
		},
		"FillAndValidateStatsDuplicate": {
			in: &Results{
				Common: &TestsConfig{
					Stats: &Stats{},
				},
				FerretDB: &TestsConfig{
					Stats: &Stats{},
				},
				MongoDB: &TestsConfig{
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

			err := c.fillAndValidate()

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

				{tc.expected.MongoDB.Pass, tc.in.MongoDB.Pass},
				{tc.expected.MongoDB.Skip, tc.in.MongoDB.Skip},
				{tc.expected.MongoDB.Fail, tc.in.MongoDB.Fail},
			} {
				//if tests.expected == nil {
				//	assert.Equal(t, tests.expected, tests.actual)
				//	continue
				//}

				//TODO: we're not checking for Default value
				for _, item := range tests.expected.TestNames {
					assert.Contains(t, tests.actual.TestNames, item)
				}
				assert.Equal(t, len(tests.expected.TestNames), len(tests.actual.TestNames))

				for _, item := range tests.expected.ResRegexp {
					assert.Contains(t, tests.actual.ResRegexp, item)
				}
				assert.Equal(t, len(tests.expected.ResRegexp), len(tests.actual.ResRegexp))
			}
		})
	}
}
