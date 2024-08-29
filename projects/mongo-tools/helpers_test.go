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
	"flag"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	uriF     = flag.String("uri", "", "MongoDB URI")
	hostURIF = flag.String("host-uri", "", "MongoDB URI from the inside Docker container")
)

func TestMain(m *testing.M) {
	flag.Parse()

	if *uriF == "" {
		log.Fatal("-uri flag is required")
	}
	if *hostURIF == "" {
		log.Fatal("-host-uri flag is required")
	}

	os.Exit(m.Run())
}

// setup returns test context and per-test client connection.
func setup(tb testing.TB) (context.Context, *mongo.Client) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(*uriF))
	require.NoError(tb, err)

	require.NoError(tb, client.Ping(ctx, nil))

	tb.Cleanup(func() {
		require.NoError(tb, client.Disconnect(ctx))
	})

	return context.Background(), client
}

// runDockerComposeCommand runs command with args inside mongosh container.
func runDockerComposeCommand(tb testing.TB, command string, args ...string) {
	tb.Helper()

	bin, err := exec.LookPath("docker")
	require.NoError(tb, err)

	args = append([]string{"compose", "run", "--rm", "mongosh", command}, args...)
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	tb.Logf("Running %s", strings.Join(cmd.Args, " "))
	err = cmd.Run()
	require.NoError(tb, err)
}

// recreateDir removes and creates directory with 0o777 permissions.
func recreateDir(tb testing.TB, dir string) {
	tb.Helper()

	err := os.RemoveAll(dir)
	require.NoError(tb, err)

	// 0o777 is typically downgraded to 0o755 by umask
	err = os.Mkdir(dir, 0o777)
	require.NoError(tb, err)

	// fix after umask
	err = os.Chmod(dir, 0o777)
	require.NoError(tb, err)
}

// getDocumentsCount returns a map of collection names and their document counts.
func getDocumentsCount(tb testing.TB, ctx context.Context, db *mongo.Database) map[string]int {
	tb.Helper()

	res := make(map[string]int)

	collections, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(tb, err)

	for _, coll := range collections {
		// It's not possible to use CountDocuments because it uses aggregation pipeline
		// that are not supported by FerretDB yet.
		var doc bson.D
		err := db.RunCommand(ctx, bson.D{
			{"count", coll},
			{"query", bson.D{}},
		}).Decode(&doc)

		require.NoError(tb, err)

		res[coll] = int(doc.Map()["n"].(int32))
	}

	return res
}

// compareDatabases checks if databases have the same collections and documents.
func compareDatabases(tb testing.TB, ctx context.Context, expected, actual *mongo.Database) {
	tb.Helper()

	collections, err := expected.ListCollectionNames(ctx, bson.D{})
	require.NoError(tb, err)
	slices.Sort(collections)

	actualCollections, err := actual.ListCollectionNames(ctx, bson.D{})
	require.NoError(tb, err)
	slices.Sort(actualCollections)

	require.Equal(tb, collections, actualCollections)

	for _, coll := range collections {
		compareCollections(tb, ctx, expected.Collection(coll), actual.Collection(coll))
	}
}

// compareCollections checks if collections have the same documents.
func compareCollections(tb testing.TB, ctx context.Context, expected, actual *mongo.Collection) {
	tb.Helper()

	expectedCur, err := expected.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
	require.NoError(tb, err)

	defer expectedCur.Close(ctx)

	actualCur, err := actual.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
	require.NoError(tb, err)

	defer actualCur.Close(ctx)

	for expectedCur.Next(ctx) {
		require.True(tb, actualCur.Next(ctx))

		var expectedDoc bson.D
		err := expectedCur.Decode(&expectedDoc)
		require.NoError(tb, err)

		var actualDoc bson.D
		err = actualCur.Decode(&actualDoc)
		require.NoError(tb, err)

		require.Equal(tb, expectedDoc, actualDoc)
	}

	require.False(tb, actualCur.Next(ctx))

	require.NoError(tb, expectedCur.Err())
	require.NoError(tb, actualCur.Err())
}
