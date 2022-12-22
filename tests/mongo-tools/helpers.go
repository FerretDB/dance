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
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setup returns test context and per-test client connection.
func setup(tb testing.TB) (context.Context, *mongo.Client) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	require.NoError(tb, err)
	err = client.Ping(ctx, nil)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
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

// hashFile returns SHA-256 hash of the file.
func hashFile(tb testing.TB, path string) string {
	tb.Helper()

	f, err := os.Open(path)
	require.NoError(tb, err)

	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	require.NoError(tb, err)

	return fmt.Sprintf("%064x", h.Sum(nil))
}

// getDirectory returns a map of file names and their hashes.
//
// Directory must not contain subdirectories.
func getDirectory(tb testing.TB, dir string) map[string]string {
	tb.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(tb, err)

	res := make(map[string]string)
	for _, entry := range entries {
		require.False(tb, entry.IsDir())

		res[entry.Name()] = hashFile(tb, filepath.Join(dir, entry.Name()))
	}

	return res
}
