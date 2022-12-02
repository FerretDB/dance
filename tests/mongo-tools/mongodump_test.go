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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/dance/tests/common"
)

func TestMongodump(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		setupFun func(ctx context.Context, db *mongo.Database)
	}{
		"Empty": {},
		"SingleDocument": {
			setupFun: func(ctx context.Context, db *mongo.Database) {
				_, err := db.Collection("mongodump").InsertOne(ctx, bson.D{{"foo", "bar"}})
				require.NoError(t, err)
			},
		},
		"ManyDocuments": {
			setupFun: func(ctx context.Context, db *mongo.Database) {
				for i := 0; i < 500; i++ {
					var value any = i

					if i%2 == 0 {
						value = fmt.Sprintf("foo%d", i)
					}

					_, err := db.Collection("mongodump").InsertOne(ctx, bson.D{{"v", value}})
					require.NoError(t, err)
				}
			},
		},
		"ManyCollections": {
			setupFun: func(ctx context.Context, db *mongo.Database) {
				for i := 0; i < 100; i++ {
					_, err := db.Collection(fmt.Sprintf("mongodump%d", i)).InsertOne(ctx, bson.D{{"v", i}})
					require.NoError(t, err)
				}
			},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runMongodumpTest(t, tc.setupFun)
		})
	}
}

// runMongodumpTest runs setupDB function which initialize a database in a specified way.
// After that it runs mongodump against this database, drop it and runs mongorestore to compare
// database state before and after restoring.
func runMongodumpTest(t *testing.T, setupDB func(context.Context, *mongo.Database)) {
	t.Helper()
	ctx, db := common.Setup(t)

	dbName := common.DatabaseName(t)

	localPath := filepath.Join("..", "..", "dumps", dbName)
	containerPath := "/dumps/" + dbName

	// set database state
	if setupDB != nil {
		setupDB(ctx, db)
	}

	// get database state before restore
	expectedState := getDatabaseState(t, ctx, db)

	// cleanup dump directory
	err := os.RemoveAll(localPath)
	require.NoError(t, err)

	// dump a database
	err = runDockerComposeCommand(
		"mongodump",
		"mongodb://host.docker.internal:27017/"+dbName,
		"-o", containerPath,
		"--verbose",
	)
	require.NoError(t, err)

	// cleanup database
	ctx, db = common.Setup(t)

	// Create directory if mongodump didn't export anything
	// It's required for mongorestore to not fail
	err = os.MkdirAll(localPath, 0o7777)
	require.NoError(t, err)

	// restore a database based on created dump
	err = runDockerComposeCommand(
		"mongorestore",
		"mongodb://host.docker.internal:27017/",
		"--dir", containerPath,
		"--verbose",
	)
	require.NoError(t, err)

	// get database state after restore
	actualState := getDatabaseState(t, ctx, db)
	assert.Equal(t, expectedState, actualState)
}
