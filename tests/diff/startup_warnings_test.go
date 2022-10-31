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

package diff

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestStartupWarnigns(t *testing.T) {
	t.Parallel()

	// TODO How to enable and disable telemetry when FerretDB is running?
	ctx, db := setup(t)

	command := bson.D{{"getLog", "startupWarnings"}}

	var actual bson.D
	err := db.RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	log := actual.Map()["log"].(bson.A)

	t.Run("FerretDB", func(t *testing.T) {
		t.Parallel()

		require.Len(t, log, 3)
		assert.Contains(t, log[2], "The telemetry state is undecided")
	})

	t.Run("MongoDB", func(t *testing.T) {
		t.Parallel()

		require.Len(t, log, 3)
		assert.NotContains(t, log[2], "The telemetry state is undecided")
	})

}
