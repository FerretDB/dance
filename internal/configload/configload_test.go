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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/dance/internal/config"
)

func FuzzLoadContent(f *testing.F) {
	for _, tc := range []struct { //nolint:vet // for readability
		file     string
		db       string
		expected *config.Config
		err      string
	}{
		{
			file: "command.yml",
			db:   "ferretdb",
			expected: &config.Config{
				Runner: "command",
				Params: &config.RunnerParamsCommand{
					Dir: "test",
					Setup: "python3 -m venv .\n" +
						"./bin/pip3 install -r requirements.txt\n",
					Tests: []config.RunnerParamsCommandTest{
						{Name: "normal", Cmd: "./bin/python3 pymongo_test.py 'mongodb://127.0.0.1:27003/'"},
						{Name: "strict", Cmd: "./bin/python3 pymongo_test.py --strict 'mongodb://127.0.0.1:27003/'"},
					},
				},
				Results: &config.ExpectedResults{
					Default: config.Pass,
					Stats: &config.Stats{
						Failed: 1,
						Passed: 1,
					},
					Fail: []string{"strict"},
				},
			},
		},
		{
			file: "command_nodir.yml",
			db:   "ferretdb",
			err:  "failed to convert runner parameters: dir is required",
		},
	} {
		b, err := os.ReadFile(filepath.Join("testdata", tc.file))
		require.NoError(f, err, "file = %s", tc.file)

		actual, err := loadContent(string(b), tc.db)
		if tc.err == "" {
			require.NoError(f, err, "file = %s", tc.file)
			require.NotNil(f, actual, "file = %s", tc.file)
			require.Equal(f, tc.expected, actual, "file = %s", tc.file)
		} else {
			require.Error(f, err, "file = %s", tc.file)
			require.EqualError(f, err, tc.err, "file = %s", tc.file)
			require.Nil(f, actual, "file = %s", tc.file)
		}

		for db := range DBs {
			f.Add(string(b), db)
		}
	}

	f.Fuzz(func(t *testing.T, content, db string) {
		_, _ = loadContent(content, db)
	})
}
