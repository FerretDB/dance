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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCore(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	collection := db.Collection("basic_types")

	s1 := bson.D{{"_id", primitive.NewObjectID()}, {"value", "1"}}
	s2 := bson.D{{"_id", primitive.NewObjectID()}, {"value", "2"}}
	i1 := bson.D{{"_id", primitive.NewObjectID()}, {"value", 1}}
	i2 := bson.D{{"_id", primitive.NewObjectID()}, {"value", 2}}
	docs := []any{
		s1, s2, i1, i2,
	}
	_, err := collection.InsertMany(ctx, docs)
	require.NoError(t, err)
}
