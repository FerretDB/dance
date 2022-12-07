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

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
//sorts them by _id and puts into a map keyed with collection names.
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

// compareFiles takes two files and checks if they have a same content.
// If they don't, it prints a short diff view.
func compareFiles(t *testing.T, file1, file2 *os.File) {
	t.Helper()
	h := sha256.New()

	_, err := io.Copy(h, file1)
	require.NoError(t, err)

	hash1 := h.Sum(nil)

	h.Reset()

	_, err = io.Copy(h, file2)
	require.NoError(t, err)

	hash2 := h.Sum(nil)

	// compare hashes of both files
	if assert.Equal(t, hash1, hash2, "Checksum of following files is different:", file1.Name(), file2.Name()) {
		return
	}

	content1, err := io.ReadAll(file1)
	require.NoError(t, err)

	content2, err := io.ReadAll(file2)
	require.NoError(t, err)

	difflib.NewMatcher(
		difflib.SplitLines(string(content1)),
		difflib.SplitLines(string(content2)),
	)
}

// compareDirs compares two directories and their files recursively.
func compareDirs(t *testing.T, dir1, dir2 string) {
	err := filepath.WalkDir(dir1, func(path string, d fs.DirEntry, err error) error {
		comparePath := strings.Replace(path, dir1, dir2, 1)

		if d.IsDir() {
			// TODO check if both contains dir
			return nil
		}

		file1, err := os.OpenFile(path, os.O_RDONLY, 0o666)
		require.NoError(t, err)

		defer file1.Close()

		file2, err := os.OpenFile(comparePath, os.O_RDONLY, 0o666)
		assert.NoError(t, err)

		defer file2.Close()

		compareFiles(t, file1, file2)
		return nil
	})

	require.NoError(t, err)
}
