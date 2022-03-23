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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAggregate(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	t.Run("SortAggregate", func(t *testing.T) {
		t.Parallel()

		collection := db.Collection(collectionName(t))

		data := []bson.D{
			{{"_id", "1"}, {"name", "Central Park Cafe"}, {"borough", "Manhattan"}},
			{{"_id", "2"}, {"name", "Rock A Feller Bar and Grill"}, {"borough", "Queens"}},
			{{"_id", "3"}, {"name", "Empire State Pub"}, {"borough", "Brooklyn"}},
			{{"_id", "4"}, {"name", "Stan's Pizzaria"}, {"borough", "Manhattan"}},
			{{"_id", "5"}, {"name", "Jane's Deli"}, {"borough", "Brooklyn"}},
		}

		for _, v := range data {
			_, err := collection.InsertOne(ctx, v)
			require.NoError(t, err)
		}

		testCases := []struct {
			name string
			sort bson.D
			IDs  []string
			err  error
		}{
			{
				name: "SortBoroughs",
				sort: bson.D{{"borough", int32(1)}},
				IDs:  []string{"3", "5", "1", "4", "2"},
			},
			{
				name: "SortBoroughsAndNames",
				sort: bson.D{{"borough", int32(1)}, {"name", int32(1)}},
				IDs:  []string{"3", "5", "1", "4", "2"},
			},
			{
				name: "SortBoroughsPassInt64Value",
				sort: bson.D{{"borough", int64(1)}},
				IDs:  []string{"3", "5", "1", "4", "2"},
			},
			{
				name: "SortBoroughsPassFloat64Value",
				sort: bson.D{{"borough", float64(1.0)}, {"name", int64(1)}},
				IDs:  []string{"3", "5", "1", "4", "2"},
			},
			{
				name: "SortBoroughsPassNegativeValue",
				sort: bson.D{{"_id", int32(-1)}},
				IDs:  []string{"5", "4", "3", "2", "1"},
			},
			{
				name: "SortBoroughsPassString",
				sort: bson.D{{"_id", "123"}},
				err: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `Illegal key in $sort specification: _id: "123"`,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			require.NotEmpty(t, tc.name)

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				if (tc.IDs == nil) == (tc.err == nil) {
					t.Fatal("exactly one of IDs or err must be set")
				}

				cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(tc.sort))

				if tc.err != nil {
					require.Error(t, err)
					require.Equal(t, tc.err, err)
					return
				}

				require.NoError(t, err)
				require.NotNil(t, cursor)

				var actual []bson.D
				require.NoError(t, cursor.All(ctx, &actual))
				var actualIDs []string
				for _, d := range actual {
					if v, ok := d.Map()["_id"]; ok {
						v, ok := v.(string)
						if !ok {
							t.Fatal("bad _id value", v)
						}

						actualIDs = append(actualIDs, v)
					}
				}

				assert.Equal(t, tc.IDs, actualIDs)
			})
		}
	})
}
