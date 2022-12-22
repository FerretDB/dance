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

	containerSourceRoot := "/dumps/mongodb-sample-databases/dump"
	localTestsRoot := filepath.Join("..", "..", "dumps", "tests")
	containerTestsRoot := "/dumps/tests"

	//nolint:vet // for readability
	type testCase struct {
		name0          string
		documentsCount map[string]int // collection name -> document count
	}

	for _, tc := range []testCase{{
		name0: "sample_geospatial",
		documentsCount: map[string]int{
			"shipwrecks": 11095,
		},
	}, {
		name0: "sample_analytics",
		documentsCount: map[string]int{
			"accounts":     1746,
			"customers":    500,
			"transactions": 1746,
		},
	}} {
		tc := tc
		t.Run(tc.name0, func(t *testing.T) {
			t.Parallel()

			name1 := tc.name0 + "_1"
			name2 := tc.name0 + "_2"

			// pre-create directory to avoid permission issues
			recreateDir(t, filepath.Join(localTestsRoot, name1))

			ctx, client := setup(t)

			db1 := client.Database(name1)
			t.Cleanup(func() { require.NoError(t, db1.Drop(ctx)) })

			db2 := client.Database(name2)
			t.Cleanup(func() { require.NoError(t, db2.Drop(ctx)) })

			// source dump -> db1
			mongorestore(t, tc.name0, containerSourceRoot, name1)
			actualCount := getDocumentsCount(t, ctx, db1)
			assert.Equal(t, tc.documentsCount, actualCount)

			// db1 -> test dump
			mongodump(t, name1, containerTestsRoot)

			// test dump -> db2
			mongorestore(t, name1, containerTestsRoot, name2)
			actualCount = getDocumentsCount(t, ctx, db2)
			assert.Equal(t, tc.documentsCount, actualCount)

			compareDatabases(t, ctx, db1, db2)

			// we can't compare bson dump files because `mongodump` queries documents in natural order,
			// but we don't support it
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
