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
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var TestQueryOPerators = func(ctx context.Context, db *mongo.Database, t *testing.T) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		collection := db.Collection(collectionName(t))
		data := getTestDBData()

		for id, v := range data {
			_, err := collection.InsertOne(ctx, bson.D{{"_id", id}, {"value", v}})
			require.NoError(t, err)
		}

		var testCases []testCase
		// testCases = append(testCases, doublesEq...)
		// testCases = append(testCases, stringCases...)
		testCases = append(testCases, arraySize...)

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
	}
}

type testCase struct {
	name string // TODO move to map key
	q    bson.D
	o    *options.FindOptions
	v    any
	IDs  []string
	err  error
}

var doublesEq = []testCase{
	{
		q:   bson.D{{"$eq", 42.13}},
		IDs: []string{"double"},
	},
	{
		q:   bson.D{{"$eq", math.Inf(-1)}},
		IDs: []string{"double-negative-infinity"},
	},
	{
		q:   bson.D{{"$eq", math.Inf(+1)}},
		IDs: []string{"double-positive-infinity"},
	},
	{
		q:   bson.D{{"$eq", math.MaxFloat64}},
		IDs: []string{"double-max"},
	},
	{
		q:   bson.D{{"$eq", math.SmallestNonzeroFloat64}},
		IDs: []string{"double-smallest"},
	},
	// $gt
	{
		q:   bson.D{{"$gt", 42.123}},
		IDs: []string{"double-max", "double-positive-infinity"},
	},
	{
		q:   bson.D{{"$gt", math.Inf(-1)}},
		IDs: []string{"double-smallest", "double", "double-max", "double-positive-infinity"},
	},
	{
		q:   bson.D{{"$gt", math.Inf(+1)}},
		IDs: nil,
	},
	{
		q:   bson.D{{"$gt", math.MaxFloat64}},
		IDs: []string{"double-positive-infinity"},
	},
	{
		q:   bson.D{{"$gt", math.SmallestNonzeroFloat64}},
		IDs: []string{"double", "double-max", "double-positive-infinity"},
	},
}

var stringCases = []testCase{
	{
		q:   bson.D{{"$eq", "foo"}},
		IDs: []string{"string"},
	},
	{
		q:   bson.D{{"$gt", "foo"}},
		IDs: []string{},
	},
}

var arraySize = []testCase{
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
