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

func TestDocumentValidation(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			doc      bson.D
			expected mongo.CommandError
		}{
			"DollarSign": {
				doc: bson.D{{"foo$", "bar"}},
				expected: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `invalid key: "foo$" (key must not contain '$' sign)`,
				},
			},
			"DotSign": {
				doc: bson.D{{"foo.bar", "baz"}},
				expected: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `invalid key: "foo.bar" (key must not contain '.' sign)`,
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

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			doc      bson.D
			expected mongo.CommandError
		}{
			"DollarSign": {
				doc: bson.D{{"$set", bson.D{{"foo$", "bar"}}}},
				expected: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `invalid key: "foo$" (key must not contain '$' sign)`,
				},
			},
			"DotSign": {
				doc: bson.D{{"$set", bson.D{{"foo", bson.D{{"bar.baz", "qaz"}}}}}},
				expected: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `invalid key: "bar.baz" (key must not contain '.' sign)`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				// TODO: make them run in parallel https://github.com/FerretDB/FerretDB/issues/1488
				//t.Parallel()

				// initiate a collection with a valid document, so we have something to update
				collection := db.Collection("update-validation-" + name)
				_, err := collection.InsertOne(ctx, bson.D{
					{"_id", "valid"},
					{"v", int32(42)},
				})
				require.NoError(t, err)

				_, err = collection.UpdateOne(ctx, bson.D{}, tc.doc)

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
