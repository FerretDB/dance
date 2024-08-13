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

	p := &config.RunnerParamsCommand{
		Setup: "exit 1",
	}
	c, err := New(p, slog.Default())
	require.NoError(t, err)

	ctx := context.Background()

	res, err := c.Run(ctx)
	require.Error(t, err)
	assert.Nil(t, res)
}
