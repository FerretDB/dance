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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/dance/internal/config"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		expected *config.Config
		err      error
	}{
		"command": {
			expected: &config.Config{
				Runner: "command",
				Params: &config.RunnerParamsCommand{
					Dir: "test",
					Setup: "python3 -m venv .\n" +
						"./bin/pip3 install -r requirements.txt\n",
					Tests: []config.RunnerParamsCommandTest{
						{Name: "normal", Cmd: "./bin/python3 pymongo_test.py 'mongodb://127.0.0.1:27001/'"},
						{Name: "strict", Cmd: "./bin/python3 pymongo_test.py --strict 'mongodb://127.0.0.1:27001/'"},
					},
				},
				Results: &config.TestConfig{
					Default: config.Pass,
					Stats: &config.Stats{
						ExpectedFail: 1,
						ExpectedPass: 1,
					},
					Fail: []string{"strict"},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := Load(filepath.Join("testdata", name+".yml"), "ferretdb-postgresql")
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
