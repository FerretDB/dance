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

package diff

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// databaseName returns valid database name for given test.
func databaseName(t *testing.T) string {
	t.Helper()

	name := strings.ToLower(t.Name())
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")

	require.Less(t, len(name), 64)
	return name
}

// collectionName returns valid collection name for given test.
func collectionName(t *testing.T) string {
	t.Helper()

	name := strings.ToLower(t.Name())
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")

	require.Less(t, len(name), 64)
	return name
}

// setup returns test context and per-test client connection and database.
func setup(t *testing.T) (context.Context, *mongo.Database) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(t, err)
	})

	db := client.Database(databaseName(t))
	err = db.Drop(context.Background())
	require.NoError(t, err)

	return context.Background(), db
}
