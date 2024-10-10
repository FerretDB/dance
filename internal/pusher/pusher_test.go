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

package pusher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/dance/internal/config"
)

func TestPusher(t *testing.T) {
	t.Parallel()

	c, err := New("mongodb://localhost:27001/dance", Logger(t))
	require.NoError(t, err)
	t.Cleanup(c.Close)

	assert.Equal(t, "dance", c.database)

	res := map[string]config.TestResult{
		"github.com/FerretDB/dance/projects/mongo-tools/TestExportImport": {},
	}
	err = c.Push(context.Background(), "mongo-tools.yml", "ferretdb-postgresql", res)
	require.NoError(t, err)
}
