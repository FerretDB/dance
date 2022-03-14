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

		// TODO keep in sync with test_db? https://github.com/FerretDB/dance/issues/43
		data := map[string]any{
			"double":                   42.13,
			"double-zero":              0.0,
			"double-max":               math.MaxFloat64,
			"double-smallest":          math.SmallestNonzeroFloat64,
			"double-positive-infinity": math.Inf(+1),
			"double-negative-infinity": math.Inf(-1),
			"double-nan":               math.NaN(),

			"string":       "foo",
			"string-empty": "",
			// "\x00",

			"document":       map[string]any{"document": int32(42)},
			"document-empty": map[string]any{},

			"array":       primitive.A{"array", int32(42)},
			"array-empty": primitive.A{},
			"array-embedded": primitive.A{
				primitive.D{
					primitive.E{Key: "age", Value: int32(1000)},
					primitive.E{Key: "document", Value: "abc"},
					primitive.E{Key: "score", Value: float32(42.13)},
				},
				primitive.D{
					primitive.E{Key: "age", Value: int32(1000)},
					primitive.E{Key: "document", Value: "def"},
					primitive.E{Key: "score", Value: float32(42.13)},
				},
				primitive.D{
					primitive.E{Key: "age", Value: int32(1002)},
					primitive.E{Key: "document", Value: "jkl"},
					primitive.E{Key: "score", Value: int32(24)},
				},
			},
			"binary":       primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
			"binary-empty": primitive.Binary{},

			// no Undefined

			"bool-false": false,
			"bool-true":  true,

			"datetime":          time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC),
			"datetime-epoch":    time.Unix(0, 0),
			"datetime-year-min": time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
			"datetime-year-max": time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC),

			"null": nil,

			"regex":       primitive.Regex{Pattern: "foo", Options: "i"},
			"regex-empty": primitive.Regex{},

			// no DBPointer
			// no JavaScript code
			// no Symbol
			// no JavaScript code w/ scope

			"int32":      int32(42),
			"int32-zero": int32(0),
			"int32-max":  int32(math.MaxInt32),
			"int32-min":  int32(math.MinInt32),

			"timestamp":   primitive.Timestamp{T: 42, I: 13},
			"timestamp-i": primitive.Timestamp{I: 1},

			"int64":      int64(42),
			"int64-zero": int64(0),
			"int64-max":  int64(math.MaxInt64),
			"int64-min":  int64(math.MinInt64),

			// no 128-bit decimal floating point (yet)

			// no Min key
			// no Max key
		}

		for id, v := range data {
			_, err := collection.InsertOne(ctx, bson.D{{"_id", id}, {"value", v}})
			require.NoError(t, err)
		}

		testCases := []struct {
			name string // TODO move to map key
			q    bson.D
			o    *options.FindOptions // options
			v    any
			IDs  []string
			err  error
		}{
			// doubles
			// $eq
			/*
				{
					bson.D{{"$eq", 42.13}},
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
			*/

			// strings
			/*
				{
					bson.D{{"$eq", "foo"}},
					[]string{"string"},
				},
				{
					bson.D{{"$gt", "foo"}},
					[]string{},
				},
			*/

			// documents
			// TODO

			// $elemMatch
			{
				name: "elemMatchWithFilter",
				q: bson.D{
					{"_id", "array-embedded"},
				},
				o: options.Find().SetProjection(bson.D{
					{"value", bson.D{{"$elemMatch", bson.D{{"score", int32(24)}}}}},
				}),
				v: bson.A{
					bson.D{
						{"age", int32(1002)},
						{"document", "jkl"},
						{"score", int32(24)},
					}},
				IDs: []string{"array-embedded"},
			},
			// arrays
			// $size
			{
				name: "SizeInt32",
				q:    bson.D{{"value", bson.D{{"$size", int32(2)}}}},
				IDs:  []string{"array"},
			},
			{
				name: "SizeInt64",
				q:    bson.D{{"value", bson.D{{"$size", int64(2)}}}},
				IDs:  []string{"array"},
			},
			{
				name: "SizeDouble",
				q:    bson.D{{"value", bson.D{{"$size", 2.0}}}},
				IDs:  []string{"array"},
			},
			{
				name: "SizeNotFound",
				q:    bson.D{{"value", bson.D{{"$size", int32(4)}}}},
				IDs:  []string{},
			},
			{
				name: "SizeZero",
				q:    bson.D{{"value", bson.D{{"$size", 0.0}}}},
				IDs:  []string{"array-empty"},
			},
			{
				name: "SizeInvalidType",
				q:    bson.D{{"value", bson.D{{"$size", bson.D{{"$gt", int32(1)}}}}}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `$size needs a number`,
				},
			},
			{
				name: "SizeNonWhole",
				q:    bson.D{{"value", bson.D{{"$size", 2.1}}}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `$size must be a whole number`,
				},
			},
			{
				name: "SizeNaN",
				q:    bson.D{{"value", bson.D{{"$size", math.NaN()}}}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `$size must be a whole number`,
				},
			},
			{
				name: "SizeInfinity",
				q:    bson.D{{"value", bson.D{{"$size", math.Inf(1)}}}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `$size must be a whole number`,
				},
			},
			{
				name: "SizeNegative",
				q:    bson.D{{"value", bson.D{{"$size", int32(-1)}}}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `$size may not be negative`,
				},
			},
			{
				q: bson.D{{"$size", int32(2)}},
				err: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `unknown top level operator: $size. If you have a field name that starts with a '$' symbol, ` +
						`consider using $getField or $setField.`,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			if tc.name == "" {
				tc.name = fmt.Sprint(tc.q)
			}

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				if (tc.IDs == nil) == (tc.err == nil) {
					t.Fatal("exactly one of IDs or err must be set")
				}

				var cursor *mongo.Cursor
				var err error
				opts := options.Find().SetSort(bson.D{{"value", 1}})
				if tc.o != nil {
					opts = tc.o
				}
				cursor, err = collection.Find(ctx, tc.q, opts)

				if tc.err != nil {
					require.Error(t, err)
					require.Equal(t, tc.err, err)
					return
				}

				require.NoError(t, err)
				require.NotNil(t, cursor)

				var expected []bson.D
				if tc.v != nil {
					expected = append(expected, bson.D{{"_id", tc.IDs[0]}, {"value", tc.v}})
				} else {
					for _, id := range tc.IDs {
						v, ok := data[id]
						require.True(t, ok)
						expected = append(expected, bson.D{{"_id", id}, {"value", v}})
					}
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
