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

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
					assertEqualError(t, tc.expected, err)
				})

				t.Run("MongoDB", func(t *testing.T) {
					require.NoError(t, err)
				})
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			filter   bson.D
			update   bson.D
			opts     options.UpdateOptions
			expected mongo.CommandError
		}{
			"NaN": {
				filter: bson.D{{"_id", "1"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
				opts:   options.UpdateOptions{},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { update: "update-NaN", ` +
						`ordered: true, $db: "testfloatvalues", updates: [ { q: { _id: "1" }, ` +
						`u: { $set: { foo: nan.0 } } } ] } with: NaN is not supported`,
				},
			},
			"NaNWithUpsert": {
				filter: bson.D{{"_id", "1"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
				opts:   options.UpdateOptions{Upsert: pointer.ToBool(true)},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { update: "update-NaNWithUpsert", ` +
						`ordered: true, $db: "testfloatvalues", updates: [ { q: { _id: "1" }, ` +
						`u: { $set: { foo: nan.0 } }, upsert: true } ] } with: NaN is not supported`,
				},
			},
			"NegativeZero": {
				filter: bson.D{{"_id", "1"}},
				update: bson.D{{"$set", bson.D{{"foo", math.Copysign(0.0, -1)}}}},
				opts:   options.UpdateOptions{},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { update: "update-NegativeZero", ` +
						`ordered: true, $db: "testfloatvalues", updates: [ { q: { _id: "1" }, ` +
						`u: { $set: { foo: -0.0 } } } ] } with: -0 is not supported`,
				},
			},
			"NegativeZeroWithUpsert": {
				filter: bson.D{{"_id", "1"}},
				update: bson.D{{"$set", bson.D{{"foo", math.Copysign(0.0, -1)}}}},
				opts:   options.UpdateOptions{Upsert: pointer.ToBool(true)},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { update: "update-NegativeZeroWithUpsert", ` +
						`ordered: true, $db: "testfloatvalues", updates: [ { q: { _id: "1" }, ` +
						`u: { $set: { foo: -0.0 } }, upsert: true } ] } with: -0 is not supported`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				coll := "update-" + name
				_, err := db.Collection(coll).InsertOne(ctx, tc.filter)
				if err != nil {
					t.Fatal(err)
				}

				_, err = db.Collection(coll).UpdateOne(ctx, tc.filter, tc.update, &tc.opts)

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

	t.Run("FindAndModify", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			filter   bson.D
			update   bson.D
			expected mongo.CommandError
		}{
			"NaN": {
				filter: bson.D{{"_id", "1"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { findAndModify: "findAndModify-NaN", ` +
						`query: { _id: "1" }, update: { $set: { foo: nan.0 } }, $db: "testfloatvalues" } with: NaN ` +
						`is not supported`,
				},
			},
			"NegativeZero": {
				filter: bson.D{{"_id", "1"}},
				update: bson.D{{"$set", bson.D{{"foo", math.Copysign(0.0, -1)}}}},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { findAndModify: "findAndModify-NegativeZero", ` +
						`query: { _id: "1" }, update: { $set: { foo: -0.0 } }, $db: "testfloatvalues" } with: -0 ` +
						`is not supported`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				coll := "findAndModify-" + name
				_, err := db.Collection(coll).InsertOne(ctx, tc.filter)
				if err != nil {
					t.Fatal(err)
				}

				// to return error
				var update any

				err = db.Collection(coll).FindOneAndUpdate(ctx, tc.filter, tc.update).Decode(&update)

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
}
