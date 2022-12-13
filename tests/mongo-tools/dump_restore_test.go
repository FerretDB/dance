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

	"github.com/FerretDB/dance/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDumpRestore(t *testing.T) {
	ctx, db := common.Setup(t)

	dbName := "sample_geospatial"
	db = db.Client().Database(dbName)

	// cleanup database
	err := db.Drop(ctx)
	require.NoError(t, err)

	localActualPath := filepath.Join("..", "..", "dumps", "actual")
	containerActualPath := filepath.Join("/dumps", "actual")

	containerExpectedPath := filepath.Join("/dumps", "expected", "dump")
	localExpectedPath := filepath.Join("..", "..", containerExpectedPath)

	containerSourcePath := filepath.Join("/dumps", "mongodb-sample-databases", "dump")
	localSourcePath := filepath.Join("..", "..", containerSourcePath)

	// pre-create directories to avoid permission issues
	err = os.Chmod(localActualPath, 0o777)
	require.NoError(t, err)
	err = os.Chmod(localExpectedPath, 0o777)
	require.NoError(t, err)

	err = os.RemoveAll(filepath.Join(localActualPath, dbName))
	require.NoError(t, err)
	err = os.RemoveAll(filepath.Join(localExpectedPath, dbName))
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(localActualPath, dbName), 0o777) // 0o777 is typically downgraded to 0o755 by umask
	require.NoError(t, err)
	err = os.Mkdir(filepath.Join(localExpectedPath, dbName), 0o777)
	require.NoError(t, err)

	err = os.Chmod(filepath.Join(localActualPath, dbName), 0o777) // fix after umask
	require.NoError(t, err)
	err = os.Chmod(filepath.Join(localExpectedPath, dbName), 0o777)
	require.NoError(t, err)

	// restore a database from a sample dump
	mongorestore(t, dbName, localSourcePath)

	expectedState := getDatabaseState(t, ctx, db)

	// "bootstrap" the expected dump from restored database
	mongodump(t, dbName, containerExpectedPath)

	// cleanup database
	require.NoError(t, db.Drop(ctx))

	// restore a database from the expected dump
	mongorestore(t, dbName, containerExpectedPath)

	// dump a database
	mongodump(t, dbName, containerActualPath)

	// get database state after restore and compare it
	actualState := getDatabaseState(t, ctx, db)
	assert.Equal(t, expectedState, actualState)

	// compare dump files. Metadata files are not compared because they
	// contain different uuid field on every dump
	compareDirs(t, filepath.Join(localExpectedPath, dbName), filepath.Join(localActualPath, dbName), `\\*.metadata.json`)
}

// mongorestore runs mongorestore utility to restore specified db from the dump
// stored in provided path on docker container.
// In case of any error it fails the test.
func mongorestore(t *testing.T, db, path string) {
	err := runDockerComposeCommand(
		"mongorestore",
		"--nsInclude", db+".*",
		"--noIndexRestore",
		"--verbose",
		"mongodb://host.docker.internal:27017/",
		path,
	)
	require.NoError(t, err)
}

// mongodump runs mongodump utility to dump specified db and stores
// it in provided path on docker container.
// In case of any error it fails the test.
func mongodump(t *testing.T, db, path string) {
	err := runDockerComposeCommand(
		"mongodump",
		"--db", db,
		"--out", path,
		"--verbose",
		"mongodb://host.docker.internal:27017/",
	)
	require.NoError(t, err)
}
