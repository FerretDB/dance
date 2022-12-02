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

package mongotools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/dance/tests/common"
)

func TestDumpRestore(t *testing.T) {
	ctx, db := common.Setup(t)

	localRoot := filepath.Join("..", "..", "dumps")
	containerRoot := "/dumps/"

	expectedState := getDatabaseState(t, ctx, db)

	// restore a database from preprepared dump
	err := runDockerComposeCommand(
		"mongorestore",
		"--dir", filepath.Join("/sample-dumps/", "sample_analytics"),
		"--db", db.Name(),
		"--verbose",
		"mongodb://host.docker.internal:27017/",
	)
	require.NoError(t, err)

	// pre-create directories to avoid permission issues
	err = os.Chmod(localRoot, 0o777)
	require.NoError(t, err)
	err = os.RemoveAll(filepath.Join(localRoot, db.Name()))
	require.NoError(t, err)
	err = os.Mkdir(filepath.Join(localRoot, db.Name()), 0o777) // 0o777 is typically downgraded to 0o755 by umask
	require.NoError(t, err)
	err = os.Chmod(filepath.Join(localRoot, db.Name()), 0o777) // fix after umask
	require.NoError(t, err)

	// dump a database
	err = runDockerComposeCommand(
		"mongodump",
		"--out", containerRoot,
		"--db", db.Name(),
		"--verbose",
		"mongodb://host.docker.internal:27017/",
	)
	require.NoError(t, err)

	// cleanup database
	ctx, db = common.Setup(t)

	// restore a database based on created dump
	err = runDockerComposeCommand(
		"mongorestore",
		"--dir", filepath.Join(containerRoot, db.Name()),
		"--db", db.Name(),
		"--verbose",
		"mongodb://host.docker.internal:27017/",
	)
	require.NoError(t, err)

	// get database state after restore
	actualState := getDatabaseState(t, ctx, db)
	assert.Equal(t, expectedState, actualState)
}
