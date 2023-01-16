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
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestFloatValues(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			doc      bson.D
			expected mongo.CommandError
		}{
			"NaN": {
				doc: bson.D{{"_id", "1"}, {"foo", math.NaN()}},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { insert: "insert-NaN", ordered: true, ` +
						`$db: "testfloatvalues", documents: [ { _id: "1", foo: nan.0 } ] } with: NaN is not supported`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := db.Collection("insert-"+name).InsertOne(ctx, tc.doc)

				t.Run("FerretDB", func(t *testing.T) {
					t.Parallel()

					assertEqualError(t, tc.expected, err)
				})

				t.Run("MongoDB", func(t *testing.T) {
					t.Parallel()

					require.NoError(t, err)
				})
			})
		}
	})

	t.Run("InsertNegativeZero", func(t *testing.T) {
		t.Parallel()

		filter := bson.D{{"_id", "1"}}
		doc := bson.D{{"_id", "1"}, {"foo", math.Copysign(0.0, -1)}}

		collection := db.Collection("insert-negative-zero")
		_, err := collection.InsertOne(ctx, doc)
		require.NoError(t, err)

		var res bson.D
		err = collection.FindOne(ctx, filter).Decode(&res)
		require.NoError(t, err)

		var actual float64
		for _, e := range res {
			if e.Key == "foo" {
				var ok bool
				actual, ok = e.Value.(float64)
				require.True(t, ok)
			}
		}

		expected := math.Copysign(0.0, +1)

		// testify require equates -0 == 0 so this passes for both -0 and 0.
		require.Equal(t, 0, actual)

		t.Run("FerretDB", func(t *testing.T) {
			t.Parallel()

			require.Equal(t, math.Signbit(expected), math.Signbit(actual))
		})

		t.Run("MongoDB", func(t *testing.T) {
			t.Parallel()

			require.Equal(t, math.Signbit(expected), math.Signbit(actual))
		})
	})

	// TODO https://github.com/FerretDB/dance/issues/266
	/*t.Run("Update", func(t *testing.T) {

	})*/

	// TODO https://github.com/FerretDB/dance/issues/266
	/*t.Run("FindAndModify", func(t *testing.T) {

	})*/

	t.Run("UpdateOne", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			filter   bson.D
			insert   bson.D
			update   bson.D
			expected mongo.CommandError
		}{
			"MulMaxFloat64": {
				filter: bson.D{{"_id", "number"}},
				insert: bson.D{{"_id", "number"}, {"v", int32(42)}},
				update: bson.D{{"$mul", bson.D{{"v", math.MaxFloat64}}}},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `update produces invalid value: { "v": +Inf }` +
						` (update operations that produce infinity values are not allowed)`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				collection := db.Collection("update-one-" + name)
				_, err := collection.InsertOne(ctx, tc.insert)
				require.NoError(t, err)

				_, err = collection.UpdateOne(ctx, tc.filter, tc.update)

				t.Run("FerretDB", func(t *testing.T) {
					t.Parallel()

					assertEqualError(t, tc.expected, err)
				})

				t.Run("MongoDB", func(t *testing.T) {
					t.Parallel()

					require.NoError(t, err)
				})
			})
		}
	})

	t.Run("UpdateOneNegativeZero", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			insert any
			update any
			id     string
		}{
			"ZeroMulNegative": {
				id:     "zero",
				insert: int32(0),
				update: float64(-1),
			},
			"NegativeMulZero": {
				id:     "negative",
				insert: int64(-1),
				update: float64(0),
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				filter := bson.D{{"_id", tc.id}}

				collection := db.Collection("update-negative-zero-" + name)
				_, err := collection.InsertOne(ctx, bson.D{{"_id", tc.id}, {"v", tc.insert}})
				require.NoError(t, err)

				_, err = collection.UpdateOne(ctx, filter, bson.D{{"$mul", bson.D{{"v", tc.update}}}})
				require.NoError(t, err)

				var res bson.D
				err = collection.FindOne(ctx, filter).Decode(&res)
				require.NoError(t, err)

				var actual float64
				for _, e := range res {
					if e.Key == "v" {
						var ok bool
						actual, ok = e.Value.(float64)
						require.True(t, ok)
					}
				}

				expected := math.Copysign(0.0, +1)

				// testify require equates -0 == 0 so this passes for both -0 and 0.
				require.Equal(t, expected, actual)

				t.Run("FerretDB", func(t *testing.T) {
					t.Parallel()

					require.Equal(t, math.Signbit(expected), math.Signbit(actual))
				})

				t.Run("MongoDB", func(t *testing.T) {
					t.Parallel()

					require.Equal(t, math.Signbit(expected), math.Signbit(actual))
				})
			})
		}
	})
}
