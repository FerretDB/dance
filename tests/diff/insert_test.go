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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestInsertDuplicateKeys(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	doc := bson.D{{"_id", "duplicate_keys"}, {"foo", "bar"}, {"foo", "baz"}}

	_, err := db.Collection("insert-duplicate-keys").InsertOne(ctx, doc)

	t.Run("FerretDB", func(t *testing.T) {
		t.Parallel()

		expected := mongo.WriteException{WriteErrors: []mongo.WriteError{{
			Index:   0,
			Code:    2,
			Message: `invalid key: "foo" (duplicate keys are not allowed)`,
		}}}

		assert.Equal(t, expected, unsetRaw(t, err))
	})

	t.Run("MongoDB", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, err)
	})
}

func TestInsertObjectIDHexString(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)
	coll := db.Collection("insert-duplicate-id")

	hex := "000102030405060708091011"

	objID, err := primitive.ObjectIDFromHex(hex)
	require.NoError(t, err)

	_, err = coll.InsertOne(ctx, bson.D{
		{"_id", objID},
	})
	require.NoError(t, err)

	_, err = coll.InsertOne(ctx, bson.D{
		{"_id", hex},
	})

	t.Run("FerretDB", func(t *testing.T) {
		t.Parallel()

		expected := mongo.WriteException{WriteErrors: []mongo.WriteError{{
			Index:   0,
			Code:    11000,
			Message: `E11000 duplicate key error collection: TestInsertObjectIDHexString.insert-duplicate-id`,
		}}}

		assert.Equal(t, expected, unsetRaw(t, err))
	})

	t.Run("MongoDB", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, err)
	})
}
