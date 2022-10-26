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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestDatabaseName(t *testing.T) {
	t.Parallel()

	collectionName := strings.Repeat("a", 10)

	t.Run("ReservedPrefix", func(t *testing.T) {
		dbName := "_ferretdb_xxx"
		ctx, db := setup(t)
		err := db.Client().Database(dbName).CreateCollection(ctx, collectionName)

		t.Run("FerretDB", func(t *testing.T) {
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid namespace: %s.%s`, dbName, collectionName),
			}
			alt := fmt.Sprintf(`Invalid namespace: %s.%s`, dbName, collectionName)
			AssertEqualAltError(t, expected, alt, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			require.NoError(t, err)
			db.Client().Database(dbName).Drop(ctx)
		})
	})

	t.Run("Dash", func(t *testing.T) {
		dbName := "ferretdb-xxx"
		ctx, db := setup(t)
		err := db.Client().Database(dbName).CreateCollection(ctx, collectionName)

		t.Run("FerretDB", func(t *testing.T) {
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid namespace: %s.%s`, dbName, collectionName),
			}
			alt := fmt.Sprintf(`Invalid namespace: %s.%s`, dbName, collectionName)
			AssertEqualAltError(t, expected, alt, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			require.NoError(t, err)
			db.Client().Database(dbName).Drop(ctx)
		})
	})
}

func TestCollectionName(t *testing.T) {
	t.Parallel()

	t.Run("Length200", func(t *testing.T) {
		collection := strings.Repeat("a", 200)
		ctx, db := setup(t)
		dbName := db.Name()
		err := db.CreateCollection(ctx, collection)

		t.Run("FerretDB", func(t *testing.T) {
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: '%s.%s'`, dbName, collection),
			}
			AssertEqualError(t, expected, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			require.NoError(t, err)
		})
	})

	t.Run("ReservedPrefix", func(t *testing.T) {
		collection := "_ferretdb_xxx"
		ctx, db := setup(t)
		dbName := db.Name()
		err := db.CreateCollection(ctx, collection)

		t.Run("FerretDB", func(t *testing.T) {
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: '%s.%s'`, dbName, collection),
			}
			AssertEqualError(t, expected, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			require.NoError(t, err)
		})
	})

	//	t.Run("Dashes", func(t *testing.T) {
	//		collection := "ferretdb-xxx"
	//		ctx, db := setup(t)
	//		err := db.CreateCollection(ctx, collection)
	//
	//		t.Run("FerretDB", func(t *testing.T) {
	//			require.NoError(t, err)
	//		})
	//
	//		t.Run("MongoDB", func(t *testing.T) {
	//			require.NoError(t, err)
	//		})
	//	})
}
