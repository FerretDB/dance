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

// TestCollectionName documents difference in reposnes:
//   * dots
//   * dashes
//   * max length in FerretDB: 255, in MongoDB: 63.
//   * FerretDB reserved prefix is _ferretdb_, MongoDB - system..
func TestCollectionName(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)
	db = db.Client().Database("admin")
	dbName := db.Name()

	t.Run("Dot", func(t *testing.T) {
		collection := "collection_name_with_a_dot."

		t.Run("FerretDB", func(t *testing.T) {
			err := db.CreateCollection(ctx, collection)
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: '%s.%s'`, dbName, collection),
			}
			AssertEqualError(t, expected, err)
			err = db.Collection(collection).Drop(ctx)
			require.NoError(t, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			_ = db.Collection(collection).Drop(ctx)
			err := db.CreateCollection(ctx, collection)
			require.NoError(t, err)
			err = db.Collection(collection).Drop(ctx)
			require.NoError(t, err)
		})
	})

	t.Run("Dash", func(t *testing.T) {
		collection := "a-dash"

		t.Run("FerretDB", func(t *testing.T) {
			err := db.CreateCollection(ctx, collection)
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: '%s.%s'`, dbName, collection),
			}
			AssertEqualError(t, expected, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			_ = db.Collection(collection).Drop(ctx)
			err := db.CreateCollection(ctx, collection)
			require.NoError(t, err)
			err = db.Collection(collection).Drop(ctx)
			require.NoError(t, err)
		})
	})

	t.Run("Length255", func(t *testing.T) {
		collection := strings.Repeat("a", 249)

		t.Run("FerretDB", func(t *testing.T) {
			err := db.CreateCollection(ctx, collection)
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: '%s.%s'`, dbName, collection),
			}
			AssertEqualError(t, expected, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			_ = db.Collection(collection).Drop(ctx)
			err := db.CreateCollection(ctx, collection)
			require.NoError(t, err)
			err = db.Collection(collection).Drop(ctx)
			require.NoError(t, err)
		})
	})

	t.Run("ReservedPrefix", func(t *testing.T) {
		collection := "_ferretdb_xxx"
		t.Run("FerretDB", func(t *testing.T) {
			err := db.CreateCollection(ctx, collection)
			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: '%s.%s'`, dbName, collection),
			}
			AssertEqualError(t, expected, err)
		})

		t.Run("MongoDB", func(t *testing.T) {
			_ = db.Collection(collection).Drop(ctx)
			err := db.CreateCollection(ctx, collection)
			require.NoError(t, err)
			err = db.Collection(collection).Drop(ctx)
			require.NoError(t, err)
		})
	})
}
