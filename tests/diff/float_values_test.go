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
					Code:    2,
					Name:    "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { _id: "1", foo: nan.0 } with: NaN is not supported`,
				},
			},
			"Inf": {
				doc: bson.D{{"_id", "1"}, {"foo", math.Copysign(0.0, -1)}},
				expected: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { _id: "1", foo: -0.0 } with: -0 is not supported`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := db.Collection("insert-"+name).InsertOne(ctx, tc.doc)

				t.Run("FerretDB", func(t *testing.T) {
					t.Parallel()

					AssertEqualError(t, tc.expected, err)
				})

				t.Run("MongoDB", func(t *testing.T) {
					t.Parallel()

					require.NoError(t, err)
				})
			})
		}
	})
}
