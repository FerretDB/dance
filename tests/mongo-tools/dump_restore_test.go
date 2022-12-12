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

	localActualPath := filepath.Join("..", "..", "dumps", "actual")
	containerActualPath := "/dumps/actual"

	localExpectedPath := filepath.Join("..", "..", "dumps", "expected")
	containerExpectedPath := "/dumps/expected"

	dbName := "sample_geospatial"
	db = db.Client().Database(dbName)

	// cleanup database
	err := db.Drop(ctx)
	require.NoError(t, err)

	// restore a database from a sample dump
	err = runDockerComposeCommand(
		"mongorestore",
		"--nsInclude", dbName+".*",
		"--verbose",
		"--uri", "mongodb://host.docker.internal:27017/",
		"--noIndexRestore",
		containerExpectedPath,
	)
	require.NoError(t, err)

	expectedState := getDatabaseState(t, ctx, db)

	// pre-create directories to avoid permission issues
	err = os.Chmod(localActualPath, 0o777)
	require.NoError(t, err)
	err = os.RemoveAll(filepath.Join(localActualPath, dbName))
	require.NoError(t, err)
	err = os.Mkdir(filepath.Join(localActualPath, dbName), 0o777) // 0o777 is typically downgraded to 0o755 by umask
	require.NoError(t, err)
	err = os.Chmod(filepath.Join(localActualPath, dbName), 0o777) // fix after umask
	require.NoError(t, err)

	// dump a database
	err = runDockerComposeCommand(
		"mongodump",
		"--out", containerActualPath,
		"--db", dbName,
		"--verbose",
		"mongodb://host.docker.internal:27017/",
	)
	require.NoError(t, err)

	// cleanup database
	err = db.Drop(ctx)
	require.NoError(t, err)

	// restore a database based on created dump
	err = runDockerComposeCommand(
		"mongorestore",
		"--nsInclude", dbName+".*",
		"--verbose",
		"mongodb://host.docker.internal:27017/",
		containerActualPath,
	)
	require.NoError(t, err)

	// get database state after restore and compare it
	actualState := getDatabaseState(t, ctx, db)
	assert.Equal(t, expectedState, actualState)

	// compare dump files. Metadata files are not compared because they
	// contain different uuid field on every dump
	compareDirs(t, filepath.Join(localExpectedPath, dbName), filepath.Join(localActualPath, dbName), `\\*.metadata.json`)
}
