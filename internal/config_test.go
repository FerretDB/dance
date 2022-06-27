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
			//TODO: no common; no ferret/mongo; common; stats
		},
		"FillAndValidate_Duplicates_Pass": {
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
		"FillAndValidate_Duplicates_Skip": {
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
		"FillAndValidate_Duplicates_Fail": {
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
		"FillAndValidate_Duplicates_All": {
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
