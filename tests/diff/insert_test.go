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
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestInsertDuplicateKeys(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	doc := bson.D{{"_id", "duplicate_keys"}, {"foo", "bar"}, {"foo", "baz"}}

	_, err := db.Collection("insert-duplicate-keys").InsertOne(ctx, doc)

	t.Run("FerretDB", func(t *testing.T) {
		t.Parallel()

		expected := mongo.CommandError{
			Code:    2,
			Name:    "BadValue",
			Message: `invalid key: "foo" (duplicate keys are not allowed)`,
		}

		assertEqualError(t, expected, err)
	})

	t.Run("MongoDB", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, err)
	})
}

func TestInsertArrays(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	doc := bson.D{{"_id", bson.A{"foo"}}}

	_, err := db.Collection("insert-arrays").InsertOne(ctx, doc)

	t.Run("FerretDB", func(t *testing.T) {
		t.Parallel()

		expected := mongo.CommandError{
			Code:    2,
			Name:    "BadValue",
			Message: `The '_id' value cannot be of type array`,
		}

		assertEqualError(t, expected, err)
	})

	t.Run("MongoDB", func(t *testing.T) {
		t.Parallel()

		expected := mongo.CommandError{
			Code:    2,
			Name:    "BadValue",
			Message: `The '_id' value cannot be of type array`,
		}

		assertEqualError(t, expected, err)
	})
}
