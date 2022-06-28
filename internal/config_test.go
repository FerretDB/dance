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
		"FillAndValidate_Filled": {
			in: &Results{
				Common: &TestsConfig{
					Pass: []string{"a", "b"},
					Skip: []string{"c", "d"},
					Fail: []string{"e", "f"},
				},
				FerretDB: &TestsConfig{
					Pass: []string{"1", "2"},
					Skip: []string{"3", "4"},
					Fail: []string{"5"},
				},
				MongoDB: &TestsConfig{
					Pass: []string{"A", "B"},
					Skip: []string{"C"},
					Fail: []string{"D", "E"},
				},
			},
			expected: &Results{
				Common: &TestsConfig{
					Pass: []string{"a", "b"},
					Skip: []string{"c", "d"},
					Fail: []string{"e", "f"},
				},
				FerretDB: &TestsConfig{
					Pass: []string{"1", "2", "a", "b"},
					Skip: []string{"3", "4", "c", "d"},
					Fail: []string{"5", "e", "f"},
				},
				MongoDB: &TestsConfig{
					Pass: []string{"A", "B", "a", "b"},
					Skip: []string{"C", "c", "d"},
					Fail: []string{"D", "E", "e", "f"},
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
					Pass: []string{"a"},
				},
				FerretDB: &TestsConfig{
					Pass: []string{"a", "b"},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"FillAndValidateDuplicatesSkip": {
			in: &Results{
				Common: &TestsConfig{
					Skip: []string{"a"},
				},
				FerretDB: &TestsConfig{
					Skip: []string{"a", "b"},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"FillAndValidateDuplicatesFail": {
			in: &Results{
				Common: &TestsConfig{
					Fail: []string{"a"},
				},
				FerretDB: &TestsConfig{
					Fail: []string{"a", "b"},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("duplicate test or prefix: \"a\""),
		},
		"FillAndValidateDuplicatesAll": {
			in: &Results{
				FerretDB: &TestsConfig{
					Pass: []string{"a"},
					Skip: []string{"a"},
					Fail: []string{"a"},
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

			for _, field := range []struct {
				expected []string
				actual   []string
			}{
				{tc.expected.FerretDB.Pass, tc.in.FerretDB.Pass},
				{tc.expected.FerretDB.Skip, tc.in.FerretDB.Skip},
				{tc.expected.FerretDB.Fail, tc.in.FerretDB.Fail},

				{tc.expected.MongoDB.Pass, tc.in.MongoDB.Pass},
				{tc.expected.MongoDB.Skip, tc.in.MongoDB.Skip},
				{tc.expected.MongoDB.Fail, tc.in.MongoDB.Fail},
			} {
				for _, item := range field.expected {
					assert.Contains(t, field.actual, item)
				}
				assert.Equal(t, len(field.expected), len(field.actual))
			}
		})
	}
}
