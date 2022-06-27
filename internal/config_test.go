package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillAndValidate(t *testing.T) {
	// NOTE: this shouldn't be Parallel I guess?

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
			//TODO: no common; no ferret/mongo; common; stats
		},
		"FillAndValidate_Duplicates": {
			in: &Results{
				Common: &TestsConfig{
					Pass: []string{"a"},
				},
				FerretDB: &TestsConfig{
					Pass: []string{"a", "b"},
				},
				MongoDB: &TestsConfig{},
			},
			expectedErr: errors.New("test \"a\" placed both in Common and FerretDB"),
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
			//TODO: len after merge - len before merge = len common

			for _, item := range tc.expected.FerretDB.Pass {
				assert.Contains(t, tc.expected.MongoDB.Pass, item)
			}
		})
	}
}

func assertCorrectlyMerged[T comparable](t *testing.T, source []T, arr1 []T, arr2 []T) {
	for _, item := range source {
		assert.Contains(t, arr1, item)
		assert.Contains(t, arr2, item)
	}
}
