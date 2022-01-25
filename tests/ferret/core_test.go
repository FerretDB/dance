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

package ferret

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCore(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	t.Run("InsertMany", func(t *testing.T) {
		t.Skip("TODO https://github.com/FerretDB/FerretDB/issues/200")

		t.Parallel()

		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()
		valid1 := bson.D{{"_id", id1}, {"value", "valid1"}}
		duplicateID := bson.D{{"_id", id1}, {"value", "duplicateID"}}
		valid2 := bson.D{{"_id", id2}, {"value", "valid2"}}
		docs := []any{valid1, duplicateID, valid2}

		t.Run("Unordered", func(t *testing.T) {
			t.Parallel()

			collection := db.Collection(collectionName(t))

			_, err := collection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
			var writeErr mongo.BulkWriteException
			assert.ErrorAs(t, err, &writeErr)
			assert.True(t, writeErr.HasErrorCode(11000))

			cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
			require.NoError(t, cursor.All(ctx, &docs))
			assert.Equal(t, []any{valid1, valid2}, docs)
		})

		t.Run("Ordered", func(t *testing.T) {
			t.Parallel()

			collection := db.Collection(collectionName(t))

			_, err := collection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(true))
			var writeErr mongo.BulkWriteException
			assert.ErrorAs(t, err, &writeErr)
			assert.True(t, writeErr.HasErrorCode(11000))

			cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
			require.NoError(t, cursor.All(ctx, &docs))
			assert.Equal(t, []any{valid1}, docs)
		})
	})

	t.Run("QueryOperators", func(t *testing.T) {
		t.Parallel()

		collection := db.Collection(collectionName(t))

		data := map[string]any{
			// doubles
			"double":                   42.123,
			"double-negative-infinity": math.Inf(-1),
			"double-positive-infinity": math.Inf(+1),
			"double-nan":               math.NaN(),
			"double-max":               math.MaxFloat64,
			"double-smallest":          math.SmallestNonzeroFloat64,

			// strings
			"string":       "foo",
			"string-empty": "",
		}

		testCases := []struct {
			query       bson.D
			expectedIDs []string
		}{
			// doubles
			// $eq
			{
				bson.D{{"$eq", 42.123}},
				[]string{"double"},
			},
			{
				bson.D{{"$eq", math.Inf(-1)}},
				[]string{"double-negative-infinity"},
			},
			{
				bson.D{{"$eq", math.Inf(+1)}},
				[]string{"double-positive-infinity"},
			},
			{
				bson.D{{"$eq", math.MaxFloat64}},
				[]string{"double-max"},
			},
			{
				bson.D{{"$eq", math.SmallestNonzeroFloat64}},
				[]string{"double-smallest"},
			},
			// $gt
			{
				bson.D{{"$gt", 42.123}},
				[]string{"double-max", "double-positive-infinity"},
			},
			{
				bson.D{{"$gt", math.Inf(-1)}},
				[]string{"double-smallest", "double", "double-max", "double-positive-infinity"},
			},
			{
				bson.D{{"$gt", math.Inf(+1)}},
				nil,
			},
			{
				bson.D{{"$gt", math.MaxFloat64}},
				[]string{"double-positive-infinity"},
			},
			{
				bson.D{{"$gt", math.SmallestNonzeroFloat64}},
				[]string{"double", "double-max", "double-positive-infinity"},
			},

			// strings
			{
				bson.D{{"$eq", "foo"}},
				[]string{"string"},
			},
			{
				bson.D{{"$gt", "foo"}},
				[]string{},
			},
		}

		for id, v := range data {
			_, err := collection.InsertOne(ctx, bson.D{{"_id", id}, {"value", v}})
			require.NoError(t, err)
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprint(tc.query), func(t *testing.T) {
				t.Parallel()

				cursor, err := collection.Find(ctx, bson.D{{"value", tc.query}}, options.Find().SetSort(bson.D{{"value", 1}}))
				require.NoError(t, err)

				var expected []bson.D
				for _, id := range tc.expectedIDs {
					v, ok := data[id]
					require.True(t, ok)
					expected = append(expected, bson.D{{"_id", id}, {"value", v}})
				}

				var actual []bson.D
				require.NoError(t, cursor.All(ctx, &actual))
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("InsertOneFindOne", func(t *testing.T) {
		t.Parallel()

		collection := db.Collection(collectionName(t))

		for name, v := range map[string]any{
			"double":    math.MaxFloat64,
			"string":    "string",
			"document":  bson.D{{"foo", "bar"}},
			"array":     bson.A{"baz", "qux"},
			"binary":    primitive.Binary{Subtype: 0xFF, Data: []byte{0x01, 0x02, 0x03}},
			"false":     false,
			"true":      true,
			"datetime":  primitive.DateTime(time.Now().UnixMilli()),
			"null":      nil,
			"regex":     primitive.Regex{Pattern: "pattern", Options: "i"},
			"int32":     int32(math.MaxInt32),
			"timestamp": primitive.Timestamp{T: 1, I: 2},
			"int64":     int64(math.MaxInt64),
		} {
			name, v := name, v
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				id := primitive.NewObjectID()
				doc := bson.D{{"_id", id}, {"value", v}}
				_, err := collection.InsertOne(ctx, doc)
				require.NoError(t, err)

				var actual any
				err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actual)
				require.NoError(t, err)
				assert.Equal(t, doc, actual)
			})
		}
	})
}
