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
)

func TestNegativeZero(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	t.Run("Insert", func(t *testing.T) {
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

		// testify require equates -0 == 0 so this passes for both -0 and 0.
		require.Equal(t, 0.0, actual)

		t.Run("FerretDB", func(t *testing.T) {
			t.Parallel()

			require.Equal(t, math.Signbit(math.Copysign(0.0, +1)), math.Signbit(actual))
		})

		t.Run("MongoDB", func(t *testing.T) {
			t.Parallel()

			require.Equal(t, math.Signbit(math.Copysign(0.0, -1)), math.Signbit(actual))
		})
	})

	t.Run("UpdateOne", func(t *testing.T) {
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

				// testify require equates -0 == 0 so this passes for both -0 and 0.
				require.Equal(t, 0.0, actual)

				t.Run("FerretDB", func(t *testing.T) {
					t.Parallel()

					require.Equal(t, math.Signbit(math.Copysign(0.0, +1)), math.Signbit(actual))
				})

				t.Run("MongoDB", func(t *testing.T) {
					t.Parallel()

					require.Equal(t, math.Signbit(math.Copysign(0.0, -1)), math.Signbit(actual))
				})
			})
		}
	})
}
