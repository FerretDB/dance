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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportImport(t *testing.T) {
	t.Parallel()

	containerSourceRoot := "/dumps/mongodb-sample-databases/"
	containerTestsRoot := "/dumps/mongoexport_tests"
	localTestsRoot := filepath.Join("..", "..", "dumps", "mongoexport_tests")

	type testCase struct {
		coll           string // target collection
		db             string // target database
		documentsCount int    // document count
	}

	for name, tc := range map[string]testCase{
		"Shipwrecks": {
			coll:           "shipwrecks",
			db:             "sample_geospatial",
			documentsCount: 11095,
		},
		"Accounts": {
			coll:           "accounts",
			db:             "sample_analytics",
			documentsCount: 1746,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbName1 := fmt.Sprintf("%s_%s_export1", tc.db, tc.coll)
			dbName2 := fmt.Sprintf("%s_%s_export2", tc.db, tc.coll)

			sourceFile := filepath.Join(containerSourceRoot, tc.db, tc.coll+".json")
			testFile := filepath.Join(containerTestsRoot, dbName1, tc.coll+".json")

			// pre-create directory to avoid permission issues
			recreateDir(t, filepath.Join(localTestsRoot, dbName1))

			ctx, client := setup(t)

			db1 := client.Database(dbName1)
			t.Cleanup(func() { require.NoError(t, db1.Drop(ctx)) })

			db2 := client.Database(dbName2)
			t.Cleanup(func() { require.NoError(t, db2.Drop(ctx)) })

			// source file -> db1
			mongoimport(t, sourceFile, dbName1, tc.coll)
			actualCount := getDocumentsCount(t, ctx, db1)
			assert.Equal(t, tc.documentsCount, actualCount[tc.coll])

			// db1 -> test file
			mongoexport(t, testFile, dbName1, tc.coll)

			// test file -> db2
			mongoimport(t, testFile, dbName2, tc.coll)
			actualCount = getDocumentsCount(t, ctx, db2)
			assert.Equal(t, tc.documentsCount, actualCount[tc.coll])

			compareDatabases(t, ctx, db1, db2)
		})
	}
}

// mongoimport imports collection from <file> file as <db>/<coll>.
func mongoimport(t *testing.T, file, db, coll string) {
	t.Helper()

	runDockerComposeCommand(
		t,
		"mongoimport",
		"--verbose=2",
		"--authenticationDatabase=admin",
		"--db="+db,
		"--collection="+coll,
		"--file="+file,
		"--drop",
		"--numInsertionWorkers=10",
		"--stopOnError",
		*uriF,
	)
}

// mongoexport exports collection from <db>/<coll> to the <file> file.
func mongoexport(t *testing.T, file, db, coll string) {
	t.Helper()

	runDockerComposeCommand(
		t,
		"mongoexport",
		"--verbose=2",
		"--authenticationDatabase=admin",
		"--db="+db,
		"--collection="+coll,
		"--out="+file,
		"--sort={x:1}",
		*uriF,
	)
}
