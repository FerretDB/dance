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

package gotest

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/dance/internal/config"
)

func Test1(t *testing.T) {
}

func TestGoTest(t *testing.T) {
	t.Parallel()

	p := &config.RunnerParamsGoTest{
		Args: []string{"-run", `Test\d+`},
	}
	c, err := New(p, slog.Default(), true)
	require.NoError(t, err)

	ctx := context.Background()

	res, err := c.Run(ctx)
	require.NoError(t, err)

	expected := map[string]config.TestResult{
		"github.com/FerretDB/dance/internal/runner/gotest/Test1": {
			Status: "pass",
			Output: "=== RUN   Test1\n" +
				"--- PASS: Test1 (0.00s)\n",
		},
	}
	assert.Equal(t, expected, res)
}
