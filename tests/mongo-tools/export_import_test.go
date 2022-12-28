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

	"github.com/stretchr/testify/require"
)

func TestExportImport(t *testing.T) {
	t.Parallel()

	localSourceRoot := "/dumps/mongodb-sample-databases/"
	localTestsRoot := filepath.Join("..", "..", "dumps", "tests")

	type testCase struct {
		coll0          string
		db0            string
		documentsCount int // document count
	}

	for _, tc := range []testCase{{
		coll0:          "shipwrecks",
		db0:            "sample_geospatial",
		documentsCount: 11095,
	}, {
		coll0:          "accounts",
		db0:            "sample_analytics",
		documentsCount: 1746,
	}} {
		tc := tc
		name0 := tc.db0 + "-" + tc.coll0
		t.Run(name0, func(t *testing.T) {

			name1 := name0 + "_dump1"
			name2 := name0 + "_dump2"

			// pre-create directory to avoid permission issues
			recreateDir(t, filepath.Join(localTestsRoot, name1))

			ctx, client := setup(t)

			db1 := client.Database(name1)
			t.Cleanup(func() { require.NoError(t, db1.Drop(ctx)) })

			db2 := client.Database(name2)
			t.Cleanup(func() { require.NoError(t, db2.Drop(ctx)) })

			path0 := filepath.Join(localSourceRoot, name0, tc.coll0+".json")
			t.Log(path0)

			// source file -> db1
			mongoimport(t, path0, tc.db0, tc.coll0)
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
		"--db="+db,
		"--collection="+coll,
		"--file="+file,
		"--drop",
		"--numInsertionWorkers=10",
		"--stopOnError",
		"mongodb://host.docker.internal:27017/",
	)
}

func mongoexport(t *testing.T, file, db, coll string) {
	t.Helper()

	runDockerComposeCommand(
		t,
		"mongoexport",
		"--verbose=2",
		"--db="+db,
		"--collection="+coll,
		"--out="+file,
		"--sort='{x:1}'",
		"mongodb://host.docker.internal:27017/",
	)
}
