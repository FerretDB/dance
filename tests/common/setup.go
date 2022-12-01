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

package common

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// databaseName returns a stable database name for that test.
func databaseName(tb testing.TB) string {
	tb.Helper()

	name := strings.ToLower(tb.Name())
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")

	require.Less(tb, len(name), 64)
	return name
}

// Setup returns test context and per-test client connection and database.
func Setup(t *testing.T) (context.Context, *mongo.Database) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
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
