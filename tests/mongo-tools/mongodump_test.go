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
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/dance/tests/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/stretchr/testify/require"
)

func TestMongodump(t *testing.T) {
	runMongodumpTest(t, func(ctx context.Context, db *mongo.Database) {
		_, err := db.Collection("mongodump").InsertOne(ctx, bson.D{{"foo", "bar"}})
		require.NoError(t, err)

		for i := 0; i < 1000; i++ {
			_, err := db.Collection("mongodump").InsertOne(ctx, bson.D{{"v", fmt.Sprintf("foo%d", i)}})
			require.NoError(t, err)
		}

	})
}

func getDatabaseState(t *testing.T, ctx context.Context, db *mongo.Database) map[string][]bson.D {
	dbState := make(map[string][]bson.D)

	doc := struct {
		Cursor struct {
			FirstBatch []struct {
				Name string `bson:"name"`
			} `bson:"firstBatch"`
		} `bson:"cursor"`
	}{}

	err := db.RunCommand(ctx, bson.D{{"listCollections", 1}}).Decode(&doc)
	require.NoError(t, err)

	var collections []string
	for _, batch := range doc.Cursor.FirstBatch {
		collections = append(collections, batch.Name)
	}

	for _, coll := range collections {
		cur, err := db.Collection(coll).Find(ctx, bson.D{{}})
		require.NoError(t, err)

		var res []bson.D
		require.NoError(t, cur.All(ctx, &res))

		dbState[coll] = res
	}
	return dbState
}

// runMongodumpTest runs setupDB function to initialize database in a specified way.
// After that it runs mongodump against database, remove it and run mongorestore to compare
// database state before and after restoring.
func runMongodumpTest(t *testing.T, setupDB func(context.Context, *mongo.Database)) {
	t.Helper()
	ctx, db := common.Setup(t)
	dbName := strings.ToLower(t.Name())

	// set database state
	setupDB(ctx, db)

	// get current database state
	expectedState := getDatabaseState(t, ctx, db)

	// cleanup dump directory
	buffer := bytes.NewBuffer([]byte{})
	err := runCommand([]string{"rm", "-f", "-r", fmt.Sprintf("dump/%s", dbName)}, buffer)
	require.NoError(t, err)

	buffer.Reset()

	// dump a database
	err = runCommand([]string{"mongodump",
		"mongodb://dance_ferretdb:27017/" + dbName,
		"--verbose",
	}, buffer)
	require.NoError(t, err)

	buffer.Reset()

	// cleanup database
	ctx, db = common.Setup(t)

	// restore a database based on created dump
	err = runCommand([]string{"mongorestore",
		"mongodb://dance_ferretdb:27017",
		"--verbose",
	}, buffer)
	require.NoError(t, err)

	actualState := getDatabaseState(t, ctx, db)
	assert.Equal(t, expectedState, actualState)
}
