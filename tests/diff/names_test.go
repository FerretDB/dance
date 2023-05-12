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
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestDatabaseName(t *testing.T) {
	t.Parallel()

	collectionName := strings.Repeat("a", 10)

	testCases := map[string]string{
		"ReservedPrefix":     "_ferretdb_xxx",
		"NonLatin":           "データベース",
		"StartingWithNumber": "1database",
	}

	for name, dbName := range testCases {
		name, dbName := name, dbName
		t.Run(name, func(t *testing.T) {
			ctx, db := setup(t)
			err := db.Client().Database(dbName).CreateCollection(ctx, collectionName)

			t.Run("FerretDB", func(t *testing.T) {
				expected := mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: fmt.Sprintf(`Invalid namespace: %s.%s`, dbName, collectionName),
				}
				alt := fmt.Sprintf(`Invalid namespace: %s.%s`, dbName, collectionName)
				assertEqualAltError(t, expected, alt, err)
			})

			t.Run("MongoDB", func(t *testing.T) {
				require.NoError(t, err)
				db.Client().Database(dbName).Drop(ctx)
			})
		})
	}
}

func TestCollectionName(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"ReservedPrefix": "_ferretdb_xxx",
		"NonUTF-8":       string([]byte{0xff, 0xfe, 0xfd}),
	}

	t.Run("CreateCollection", func(t *testing.T) {

		for name, collection := range testCases {
			name, collection := name, collection

			t.Run(name, func(t *testing.T) {
				ctx, db := setup(t)
				dbName := db.Name()
				err := db.CreateCollection(ctx, collection)

				t.Run("FerretDB", func(t *testing.T) {
					expected := mongo.CommandError{
						Name:    "InvalidNamespace",
						Code:    73,
						Message: fmt.Sprintf(`Invalid collection name: '%s.%s'`, dbName, collection),
					}
					assertEqualError(t, expected, err)
				})

				t.Run("MongoDB", func(t *testing.T) {
					require.NoError(t, err)
				})
			})
		}
	})

	t.Run("RenameCollection", func(t *testing.T) {
		for name, collection := range testCases {
			name, collection := name, collection
			t.Run(name, func(t *testing.T) {
				ctx, db := setup(t)
				dbName := db.Name()
				collectionToCreate := "collectionToRename" + name

				err := db.CreateCollection(ctx, collectionToCreate)
				require.NoError(t, err)

				renameCommand := bson.D{
					{"renameCollection", dbName + "." + collectionToCreate},
					{"to", dbName + "." + collection},
				}
				var res bson.D
				err = db.Client().Database("admin").RunCommand(ctx, renameCommand).Decode(&res)

				t.Run("FerretDB", func(t *testing.T) {
					expected := mongo.CommandError{
						Name:    "IllegalOperation",
						Code:    20,
						Message: fmt.Sprintf(`error with target namespace: Invalid collection name: %s`, collection),
					}
					assertEqualError(t, expected, err)
				})

				t.Run("MongoDB", func(t *testing.T) {
					require.NoError(t, err)
				})
			})
		}
	})
}
