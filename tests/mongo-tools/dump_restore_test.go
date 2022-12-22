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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDumpRestore(t *testing.T) {
	t.Parallel()

	containerDump0Path := "/dumps/mongodb-sample-databases/dump"
	localDump1Path := filepath.Join("..", "..", "dumps", "dump1")
	containerDump1Path := "/dumps/dump1"
	localDump2Path := filepath.Join("..", "..", "dumps", "dump2")
	containerDump2Path := "/dumps/dump2"

	type testCase struct {
		db0            string
		documentsCount map[string]int // collection name -> document count
	}

	for _, tc := range []testCase{{
		db0: "sample_geospatial",
		documentsCount: map[string]int{
			"shipwrecks": 11095,
		},
	}, {
		db0: "sample_analytics",
		documentsCount: map[string]int{
			"accounts":     1746,
			"customers":    500,
			"transactions": 1746,
		},
	}} {
		tc := tc
		t.Run(tc.db0, func(t *testing.T) {
			t.Parallel()

			// pre-create directories to avoid permission issues
			db1 := tc.db0 + "_dump1"
			db2 := tc.db0 + "_dump2"
			recreateDir(t, filepath.Join(localDump1Path, db1))
			recreateDir(t, filepath.Join(localDump2Path, db2))

			ctx, client := setup(t)

			// dump0 -> db1 -> dump1
			t.Run("dump1", func(t *testing.T) {
				db := client.Database(db1)
				require.NoError(t, db.Drop(ctx))
				t.Cleanup(func() { require.NoError(t, db.Drop(ctx)) })

				mongorestore(t, tc.db0, containerDump0Path, db1)
				actualCount := getDocumentsCount(t, ctx, db)
				assert.Equal(t, tc.documentsCount, actualCount)

				mongodump(t, db1, containerDump1Path)

				// we can't compare dump1 files because dump0 was made with an older tool
			})

			// dump1 -> db2 -> dump2
			t.Run("dump2", func(t *testing.T) {
				db := client.Database(db2)
				require.NoError(t, db.Drop(ctx))
				t.Cleanup(func() { require.NoError(t, db.Drop(ctx)) })

				mongorestore(t, db1, containerDump1Path, db2)
				actualCount := getDocumentsCount(t, ctx, db)
				assert.Equal(t, tc.documentsCount, actualCount)

				mongodump(t, db2, containerDump2Path)

				// now we can
				expectedDir := getDirectory(t, filepath.Join(localDump1Path, db1))
				actualDir := getDirectory(t, filepath.Join(localDump2Path, db2))
				assert.Equal(t, expectedDir, actualDir)
			})
		})
	}
}

// mongorestore restores database <db> from <root>/<db> directory as <newDB>.
func mongorestore(t *testing.T, db, root, newDB string) {
	t.Helper()

	runDockerComposeCommand(
		t,
		"mongorestore",
		"--verbose=2",
		"--nsInclude="+db+".*",
		"--nsFrom="+db+".*",
		"--nsTo="+newDB+".*",
		"--objcheck",
		"--drop",
		"--noIndexRestore", // not supported by FerretDB yet
		"--numParallelCollections=10",
		"--numInsertionWorkersPerCollection=10",
		"--stopOnError",
		// "--preserveUUID", TODO https://github.com/FerretDB/FerretDB/issues/1682
		"mongodb://host.docker.internal:27017/",
		root,
	)
}

// mongodump dumps database <db> into <root>/<db> directory.
func mongodump(t *testing.T, db, root string) {
	t.Helper()

	runDockerComposeCommand(
		t,
		"mongodump",
		"--verbose=2",
		"--db="+db,
		"--out="+root,
		"--numParallelCollections=10",
		"mongodb://host.docker.internal:27017/",
	)
}
