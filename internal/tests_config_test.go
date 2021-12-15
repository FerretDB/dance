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
