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
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/dance/tests/common"
)

func TestDumpRestore(t *testing.T) {
	ctx, db := common.Setup(t)

	localRoot := filepath.Join("..", "..", "dumps")
	containerRoot := "/dumps/"

	dbName := "sample_geospatial"
	db = db.Client().Database(dbName)

	// cleanup database
	err := db.Drop(ctx)
	require.NoError(t, err)

	// restore a database from preprepared dump
	err = runDockerComposeCommand(
		"mongorestore",
		//"--nsFrom="+dbName+".*",
		//"--nsTo="+dbName+".*",
		"--nsInclude", dbName+".*",
		"--verbose",
		"--uri", "mongodb://host.docker.internal:27017/",
		"--noIndexRestore",
		filepath.Join("/sample-dumps/"),
	)
	require.NoError(t, err)

	expectedState := getDatabaseState(t, ctx, db)

	// pre-create directories to avoid permission issues
	err = os.Chmod(localRoot, 0o777)
	require.NoError(t, err)
	err = os.RemoveAll(filepath.Join(localRoot, dbName))
	require.NoError(t, err)
	err = os.Mkdir(filepath.Join(localRoot, dbName), 0o777) // 0o777 is typically downgraded to 0o755 by umask
	require.NoError(t, err)
	err = os.Chmod(filepath.Join(localRoot, dbName), 0o777) // fix after umask
	require.NoError(t, err)

	// dump a database
	err = runDockerComposeCommand(
		"mongodump",
		"--out", containerRoot,
		"--db", dbName,
		"--verbose",
		"mongodb://host.docker.internal:27017/",
	)
	require.NoError(t, err)

	// cleanup database
	err = db.Drop(ctx)
	require.NoError(t, err)

	//// restore a database based on created dump
	err = runDockerComposeCommand(
		"mongorestore",
		"--nsInclude", dbName+".*",
		"--verbose",
		"mongodb://host.docker.internal:27017/",
		filepath.Join(containerRoot),
	)
	require.NoError(t, err)

	// get database state after restore
	actualState := getDatabaseState(t, ctx, db)
	assert.Equal(t, len(expectedState), len(actualState))

	compareDatabaseStates(t, expectedState, actualState)

	compareDirs(t, filepath.Join("..", "..", "sample-dump", dbName), filepath.Join(localRoot, db.Name()))
}

// #
// - data restored from mongodb is the same in ferretdb
//   - NOTE: the order of documents may be different
//   - because of that, the dump file differs
//
// - dump file from ferretdb is the same every time on the same database
//   - that would prove that dumping/restoring process is correct

// To make sure that above requirements are met we want to (step by step):
// - restore the data from the prepared dump
// - get the current database state (from mongo driver)
// - dump the data
// - drop the database and restore from a last dump
// - compare current state of database with the previous one <- proves that the restoring and dumping process works **on restored data from prepared dump that COULD be invalid**
// - compare old dump file with a new one <- proves that the first restore was for sure valid (it didn't omit any data)

// TODO: make a note about 3yo json dumps

// if we need to compare mongodb data to ferretdb data we need to:
// (let's remember that dump file from mongodb is a source of truth in case of data consistency)
// - restore prepared dump
// - dump it again
// - compare dumps (if they're the same, the data is either
//
// comparing data with mongo driver after the first restore and the second one is not sufficient, because
// it will allow data inconcistency on the first dump and compare the INVALID data with the second dump
//
// The above approach is not possible, due to ferretdb documents order (which is for now different than mongodb and I'm not sure do we have fixing that in plans)
// - to resolve this issue, we need another source of truth for FerretDB
// - we know that FerretDB dump will store documents in the same order
// - so if we somehow prepare FerretDB dump that is valid with mongodb dump, we are sure that everything is proceed correctly
// - this, is possible by manually running tests on mongodb dump and ferretdb dump of the mongodb dump and storing both of them in sample directory.
//   - the data is restored from mongodb dump
//   - expectedState value is set to database state after mongodb dump
//   - we clear the db
//   - the data is restored from ferretdb dump
//   - actualState value is set to database state after ferretdb dump
//   - we compare both of states
//   - they are not the same because of different order of documents in collections
//   - to prove above sentence the test is running inefficient o(n2) loops that checks if the specified document occures in a same collection at the same index
//     - if it is, then we are sure that the order for specified documents sequence is the same
//	   - if it has a different index then we now that the document exists in collection but the order is not the same as in mongo,
//     - if one of above statements is correct, that means that the data in ferretdb dump is the same as in mongodb and it differs ONLY
//     by an order.

// First step: seed the data
// - use db specific dump file

// First step: seed the data
// - Find an existent dump file [a b c]
// - Call restore for that file [a b c] != [a x c]
// Result: DB state seed [ a x c ]
// -----
// Second step: check if the dumps of FerretDB itself are idempotent
// Dump1
// Restore1
// Result: DB state 1
// Dump2
// Restore2
// Result: DB state 2
// Compare Dump1 == Dump2
// Compare DB state 1 == DB state 2
// DB state seed == DB state 1 ???
