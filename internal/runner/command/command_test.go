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

package command

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/dance/internal/config"
)

func TestCommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	p := &config.RunnerParamsCommand{
		Tests: []config.RunnerParamsCommandTest{
			{
				Name: "test1",
				Cmd:  "exit 0",
			},
			{
				Name: "test2",
				Cmd:  "exit 1",
			},
		},
	}

	t.Run("Normal", func(t *testing.T) {
		c, err := New(p, slog.Default(), false)
		require.NoError(t, err)

		res, err := c.Run(ctx)
		require.NoError(t, err)

		expected := map[string]config.TestResult{
			"test1": {
				Status: "pass",
				Output: "",
			},
			"test2": {
				Status: "fail",
				Output: "\nexit status 1",
			},
		}
		assert.Equal(t, expected, res)
	})

	t.Run("Verbose", func(t *testing.T) {
		c, err := New(p, slog.Default(), true)
		require.NoError(t, err)

		res, err := c.Run(ctx)
		require.NoError(t, err)

		expected := map[string]config.TestResult{
			"test1": {
				Status: "pass",
				Output: "+ exit 0\n",
			},
			"test2": {
				Status: "fail",
				Output: "+ exit 1\n\n" +
					"exit status 1",
			},
		}
		assert.Equal(t, expected, res)
	})
}
