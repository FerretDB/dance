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
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// databaseName returns a stable database name for that test.
func databaseName(tb testing.TB) string {
	tb.Helper()

	// database names are always lowercase
	name := strings.ToLower(tb.Name())

	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "$", "_")

	require.Less(tb, len(name), 64)
	return name
}

// setup returns test context and per-test client connection and database.
func setup(t *testing.T) (context.Context, *mongo.Database) {
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

// runDockerComposeCommand runs command with args inside mongosh container.
func runDockerComposeCommand(command string, args ...string) error {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	dockerArgs := append([]string{"compose", "run", "--rm", "mongosh", command}, args...)
	cmd := exec.Command(bin, dockerArgs...)
	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(dockerArgs, " "), err)
	}

	return nil
}

// getDatabaseState gets all documents from each collection in provided database,
// sorts them by _id and puts into a map keyed with collection names.
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
		cur, err := db.Collection(coll).Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
		require.NoError(t, err)

		var res []bson.D
		require.NoError(t, cur.All(ctx, &res))

		dbState[coll] = res
	}

	return dbState
}

// compareFiles takes two file paths and checks if they have the same content.
func compareFiles(t *testing.T, path, comparePath string) {
	t.Helper()
	h := sha256.New()

	file1, err := os.Open(path)
	require.NoError(t, err)

	defer file1.Close()

	if !assert.FileExists(t, comparePath) {
		return
	}

	file2, err := os.Open(comparePath)
	require.NoError(t, err)

	defer file2.Close()

	_, err = io.Copy(h, file1)
	require.NoError(t, err)

	hash1 := h.Sum(nil)

	h.Reset()

	_, err = io.Copy(h, file2)
	require.NoError(t, err)

	hash2 := h.Sum(nil)

	// compare hashes of both files
	if assert.Equal(t, hash1, hash2, "Checksums of following files are different:", file1.Name(), file2.Name()) {
		return
	}

	// reset file offsets and compares the content of both of them to show
	// a more detailed output

	_, err = file1.Seek(0, io.SeekStart)
	require.NoError(t, err)

	_, err = file2.Seek(0, io.SeekStart)
	require.NoError(t, err)

	content1, err := io.ReadAll(file1)
	require.NoError(t, err)

	content2, err := io.ReadAll(file2)
	require.NoError(t, err)

	require.Equal(t, content1, content2)
}

// compareDirs compares two directories and their files recursively.
// It ignores files based on globs provided in ignoredFiles.
func compareDirs(t *testing.T, dir1, dir2 string, ignoredFiles ...string) {
	err := filepath.WalkDir(dir1, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		comparePath := strings.Replace(path, dir1, dir2, 1)

		// skip all ignored paths
		for _, pattern := range ignoredFiles {
			var ignore bool
			ignore, err = filepath.Match(pattern, d.Name())
			require.NoError(t, err)

			if ignore {
				t.Logf("Skipping comparison of %s", path)
				return nil
			}
		}

		if d.IsDir() {
			_, err = os.Stat(comparePath)
			assert.NoError(t, err)
			return nil
		}

		compareFiles(t, path, comparePath)
		return nil
	})
	require.NoError(t, err)
}
