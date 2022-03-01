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

	t.Run("Limit", func(t *testing.T) {
		t.Parallel()

		collection := db.Collection(collectionName(t))

		docs := []any{
			bson.D{{"_id", "1"}},
			bson.D{{"_id", "2"}},
			bson.D{{"_id", "3"}},
		}
		_, err := collection.InsertMany(ctx, docs)
		require.NoError(t, err)

		cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetLimit(1).SetSort(bson.D{{"_id", 1}}))
		require.NoError(t, err)

		var actual []any
		require.NoError(t, cursor.All(ctx, &actual))
		assert.Equal(t, []any{docs[0]}, actual)
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

			"document":       bson.D{{"document", int32(42)}},
			"document-empty": bson.D{},

			"array":       primitive.A{"array", int32(42)},
			"array-three": primitive.A{int32(42), "foo", nil},
			"array-empty": primitive.A{},
			"array-embedded": bson.A{
				bson.D{{"age", 1000}, {"document", "abc"}, {"score", 42.13}},
				bson.D{{"age", 1000}, {"document", "def"}, {"score", 42.13}},
				bson.D{{"age", 1002}, {"document", "jkl"}, {"score", 24}},
			},

			"binary":       primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
			"binary-empty": primitive.Binary{Data: []byte{}},
			"binary-big":   primitive.Binary{Data: []byte{0, 0, 128}},

			// no Undefined

			"bool-false": false,
			"bool-true":  true,

			"datetime":          primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)),
			"datetime-epoch":    primitive.NewDateTimeFromTime(time.Unix(0, 0)),
			"datetime-year-min": primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)),
			"datetime-year-max": primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC)),

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

		// It's being checked only one of (v, IDs, err).
		testCases := []struct {
			name string // TODO move to map key
			q    bson.D
			o    *options.FindOptions // if empty, defaults to sorting by value for stable tests
			v    []bson.D             // expected value; useful for testing projections
			IDs  []string             // expected values IDs; useful when projections are not used
			err  error
		}{
			// doubles
			// $eq
			{
				name: "EqDouble",
				q:    bson.D{{"value", bson.D{{"$eq", 42.13}}}},
				IDs:  []string{"double"},
			},
			{
				name: "EqDoubleNegativeInfinity",
				q:    bson.D{{"value", bson.D{{"$eq", math.Inf(-1)}}}},
				IDs:  []string{"double-negative-infinity"},
			},
			{
				name: "EqDoublePositiveInfinity",
				q:    bson.D{{"value", bson.D{{"$eq", math.Inf(+1)}}}},
				IDs:  []string{"double-positive-infinity"},
			},
			{
				name: "EqDoubleMax",
				q:    bson.D{{"value", bson.D{{"$eq", math.MaxFloat64}}}},
				IDs:  []string{"double-max"},
			},
			{
				name: "EqDoubleSmallest",
				q:    bson.D{{"value", bson.D{{"$eq", math.SmallestNonzeroFloat64}}}},
				IDs:  []string{"double-smallest"},
			},
			{
				name: "EqDoubleZero",
				q:    bson.D{{"value", bson.D{{"$eq", 0.0}}}},
				IDs:  []string{"double-zero"},
			},

			// $gt
			/*

					{
					q:bson.D{{"$gt", 42.123}},
					IDs:[]string{"double-max", "double-positive-infinity"},
				},
				{
					q:bson.D{{"$gt", math.Inf(-1)}},
					IDs:[]string{"double-smallest", "double", "double-max", "double-positive-infinity"},
				},
				{
					q:bson.D{{"$gt", math.Inf(+1)}},
					nil,
				},
				{
					q:bson.D{{"$gt", math.MaxFloat64}},
					IDs:[]string{"double-positive-infinity"},
				},
				{
					q:bson.D{{"$gt", math.SmallestNonzeroFloat64}},
					IDs:[]string{"double", "double-max", "double-positive-infinity"},
				},
			*/

			// strings

			{
				name: "FindManyString",
				q:    bson.D{{"value", "foo"}},
				o:    options.Find().SetSort(bson.D{{"_id", 1}}),
				IDs:  []string{"array-three", "string"},
			},
			//{
			//	bson.D{{"$gt", "foo"}},
			//	[]string{},
			//},

			// int32

			{
				name: "FindManyInt32",
				q:    bson.D{{"value", int32(42)}},
				o:    options.Find().SetSort(bson.D{{"_id", 1}}),
				IDs:  []string{"array", "array-three", "int32", "int64"},
			},

			// int64

			{
				name: "FindManyInt64",
				q:    bson.D{{"value", int64(42)}},
				o:    options.Find().SetSort(bson.D{{"_id", 1}}),
				IDs:  []string{"array", "array-three", "int32", "int64"},
			},

			//  $eq

			{
				name: "EqString",
				q:    bson.D{{"value", bson.D{{"$eq", "foo"}}}},
				IDs:  []string{"string"},
			},
			{
				name: "EqEmptyString",
				q:    bson.D{{"value", bson.D{{"$eq", ""}}}},
				IDs:  []string{"string-empty"},
			},

			// documents
			// TODO

			// projection
			{
				name: "ProjectionInclusion",
				q: bson.D{
					{"_id", "double"},
				},
				o:   options.Find().SetProjection(bson.D{{"value", int32(11)}}),
				IDs: []string{"double"},
			},
			{
				name: "ProjectionExclusion",
				q: bson.D{
					{"_id", "double"},
				},
				o: options.Find().SetProjection(bson.D{{"value", false}}),
				v: []bson.D{{{"_id", "double"}}},
			},
			{
				name: "ProjectionBothErrorInclusion",
				q: bson.D{
					{"_id", "document-diverse"},
				},
				o: options.Find().SetProjection(bson.D{
					{"document_id", false},
					{"array", true},
				}),
				err: mongo.CommandError{
					Code:    31253,
					Name:    "Location31253",
					Message: `Cannot do inclusion on field array in exclusion projection`,
				},
			},
			{
				name: "ProjectionBothErrorExclusion",
				q: bson.D{
					{"_id", "document-diverse"},
				},
				o: options.Find().SetProjection(bson.D{
					{"array", true},
					{"document_id", false},
				}),
				err: mongo.CommandError{
					Code:    31254,
					Name:    "Location31254",
					Message: `Cannot do exclusion on field document_id in inclusion projection`,
				},
			},
			{
				name: "ProjectionElemMatchWithFilter",
				q:    bson.D{{"_id", "array-embedded"}},
				o: options.Find().SetProjection(bson.D{
					{"value", bson.D{{"$elemMatch", bson.D{{"score", int32(24)}}}}},
				}),
				v: []bson.D{{
					{"_id", "array-embedded"},
					{"value", bson.A{
						bson.D{
							{"age", int32(1002)},
							{"document", "jkl"},
							{"score", int32(24)},
						},
					}},
				}},
			},

			{
				name: "BitsAllClear",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllClear", int32(21)}}}},
				IDs:  []string{"int32"},
			},
			{
				name: "BitsAllClearEmptyResult",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllClear", int32(53)}}}},
				IDs:  []string{},
			},
			{
				name: "BitsAllClearString",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllClear", "123"}}}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: "value takes an Array, a number, or a BinData but received: $bitsAllClear: \"123\"",
				},
			},
			{
				name: "BitsAllClearPassFloat",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllClear", 1.2}}}},
				err: mongo.CommandError{
					Code:    9,
					Name:    "FailedToParse",
					Message: "Expected an integer: $bitsAllClear: 1.2",
				},
			},
			{
				name: "BitsAllClearPassNegativeValue",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllClear", int32(-1)}}}},
				err: mongo.CommandError{
					Code:    9,
					Name:    "FailedToParse",
					Message: "Expected a positive number in: $bitsAllClear: -1",
				},
			},
			{
				name: "BitsAllSet",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllSet", int32(42)}}}},
				IDs:  []string{"int32"},
			},
			{
				name: "BitsAllSetEmpty",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllSet", int32(43)}}}},
				IDs:  []string{},
			},
			{
				name: "BitsAllSetString",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllSet", "123"}}}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: "value takes an Array, a number, or a BinData but received: $bitsAllSet: \"123\"",
				},
			},
			{
				name: "BitsAllSetPassFloat",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllSet", 1.2}}}},
				err: mongo.CommandError{
					Code:    9,
					Name:    "FailedToParse",
					Message: "Expected an integer: $bitsAllSet: 1.2",
				},
			},
			{
				name: "BitsAllSetPassNegativeValue",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAllSet", int32(-1)}}}},
				err: mongo.CommandError{
					Code:    9,
					Name:    "FailedToParse",
					Message: "Expected a positive number in: $bitsAllSet: -1",
				},
			},
			{
				name: "BitsAnyClear",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAnyClear", int32(1)}}}},
				IDs:  []string{"int32"},
			},
			{
				name: "BitsAnyClearEmpty",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAnyClear", int32(42)}}}},
				IDs:  []string{},
			},
			{
				name: "BitsAnyClearBigBinary",
				q: bson.D{{"_id", "binary-big"}, {
					"value",
					bson.D{{"$bitsAnyClear", int32(0b1000_0000_0000_0000)}},
				}},
				IDs: []string{"binary-big"},
			},
			{
				name: "BitsAnyClearBigBinaryEmptyResult",
				q: bson.D{{"_id", "binary-big"}, {
					"value",
					bson.D{{"$bitsAnyClear", int32(0b1000_0000_0000_0000_0000_0000)}},
				}},
				IDs: []string{},
			},
			{
				name: "BitsAnySet",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAnySet", int32(2)}}}},
				IDs:  []string{"int32"},
			},
			{
				name: "BitsAnySetEmpty",
				q:    bson.D{{"_id", "int32"}, {"value", bson.D{{"$bitsAnySet", int32(4)}}}},
				IDs:  []string{},
			},
			{
				name: "BitsAnySetBigBinary",
				q: bson.D{{"_id", "binary-big"}, {
					"value",
					bson.D{{"$bitsAnySet", int32(0b1000_0000_0000_0000_0000_0000)}},
				}},
				IDs: []string{"binary-big"},
			},
			{
				name: "BitsAnySetBigBinaryEmptyResult",
				q: bson.D{{"_id", "binary-big"}, {
					"value",
					bson.D{{"$bitsAnySet", int32(0b1000_0000_0000_0000)}},
				}},
				IDs: []string{},
			},

			// binary

			// $eq
			{
				name: "EqBinary",
				q:    bson.D{{"value", bson.D{{"$eq", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
				IDs:  []string{"binary"},
			},

			// bool

			// $eq

			{
				name: "EqBoolTrue",
				q:    bson.D{{"value", bson.D{{"$eq", true}}}},
				IDs:  []string{"bool-true"},
			},
			{
				name: "EqBoolFalse",
				q:    bson.D{{"value", bson.D{{"$eq", false}}}},
				IDs:  []string{"bool-false"},
			},

			// datetime

			// $eq

			{
				name: "EqDatetime",
				q:    bson.D{{"value", bson.D{{"$eq", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)}}}},
				IDs:  []string{"datetime"},
			},

			// timestamp

			// $eq

			{
				name: "EqTimestamp",
				q:    bson.D{{"value", bson.D{{"$eq", primitive.Timestamp{T: 42, I: 13}}}}},
				IDs:  []string{"timestamp"},
			},

			// null

			{
				name: "EqNull",
				q:    bson.D{{"value", bson.D{{"$eq", nil}}}},
				IDs:  []string{"null"},
			},

			// regex

			{
				name: "FindRegex",
				q:    bson.D{{"value", primitive.Regex{Pattern: "foo", Options: "i"}}},
				IDs:  []string{"string", "regex"},
			},

			// $eq

			{
				name: "EqRegex",
				q:    bson.D{{"value", bson.D{{"$eq", primitive.Regex{Pattern: "foo", Options: "i"}}}}},
				IDs:  []string{"regex"},
			},

			// int

			// $eq

			{
				name: "EqInt64",
				q:    bson.D{{"value", bson.D{{"$eq", int64(42)}}}},
				IDs:  []string{"int64"},
			},
			{
				name: "EqIntZero",
				q:    bson.D{{"value", bson.D{{"$eq", int64(0)}}}},
				IDs:  []string{"int64-zero"},
			},
			{
				name: "EqIntMax",
				q:    bson.D{{"value", bson.D{{"$eq", int64(math.MaxInt64)}}}},
				IDs:  []string{"int64-max"},
			},
			{
				name: "EqIntMin",
				q:    bson.D{{"value", bson.D{{"$eq", int64(math.MinInt64)}}}},
				IDs:  []string{"int64-min"},
			},
		}

		t.Run("EqNaN", func(t *testing.T) {
			t.Parallel()

			var result bson.M
			err := collection.FindOne(ctx, bson.D{{"value", bson.D{{"$eq", math.NaN()}}}}).Decode(&result)
			require.NoError(t, err)
			if nan, ok := result["value"].(float64); ok {
				assert.True(t, math.IsNaN(nan))
			}
		})

		for _, tc := range testCases {
			tc := tc

			if tc.name == "" {
				tc.name = fmt.Sprint(tc.q)
			}

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				if (tc.IDs == nil) == (tc.err == nil) == (tc.v == nil) {
					t.Fatal("exactly one of IDs, err or v must be set")
				}

				var cursor *mongo.Cursor
				var err error
				if tc.o == nil {
					tc.o = options.Find().SetSort(bson.D{{"value", 1}})
				}
				cursor, err = collection.Find(ctx, tc.q, tc.o)

				if tc.err != nil {
					require.Error(t, err)
					require.Equal(t, tc.err, err)
					return
				}

				require.NoError(t, err)
				require.NotNil(t, cursor)

				var expected []bson.D
				if tc.v != nil {
					expected = tc.v
				} else {
					for _, id := range tc.IDs {
						v, ok := data[id]
						require.True(t, ok)
						expected = append(expected, bson.D{{"_id", id}, {"value", v}})
					}
				}

				var actual []bson.D
				require.NoError(t, cursor.All(ctx, &actual))
				if !assert.Equal(t, expected, actual) {
					// a diff of IDs is easier to read
					expectedIDs := make([]string, len(expected))
					for i, e := range expected {
						expectedIDs[i] = e.Map()["_id"].(string)
					}
					actualIDs := make([]string, len(actual))
					for i, a := range actual {
						actualIDs[i] = a.Map()["_id"].(string)
					}
					assert.Equal(t, expectedIDs, actualIDs)
				}
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

	t.Run("MostCommandsAreCaseSensitive", func(t *testing.T) {
		t.Parallel()

		res := db.RunCommand(ctx, bson.D{{"listcollections", 1}})
		err := res.Err()
		require.Error(t, err)
		assert.Equal(t, mongo.CommandError{Code: 59, Name: "CommandNotFound", Message: `no such command: 'listcollections'`}, err)

		res = db.RunCommand(ctx, bson.D{{"listCollections", 1}})
		assert.NoError(t, res.Err())

		// special case
		res = db.RunCommand(ctx, bson.D{{"ismaster", 1}})
		assert.NoError(t, res.Err())
		res = db.RunCommand(ctx, bson.D{{"isMaster", 1}})
		assert.NoError(t, res.Err())
	})
}
