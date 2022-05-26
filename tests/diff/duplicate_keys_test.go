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

	id1 := primitive.NewObjectID()
	id2 := primitive.NewObjectID()
	valid1 := bson.D{{"_id", id1}, {"value", "valid1"}}
	duplicateID := bson.D{{"_id", id1}, {"value", "duplicateID"}}
	valid2 := bson.D{{"_id", id2}, {"value", "valid2"}}
	docs := []any{valid1, duplicateID, valid2}

	collection := db.Collection(collectionName(t))

	_, err := collection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	var writeErr mongo.BulkWriteException
	assert.ErrorAs(t, err, &writeErr)
	assert.True(t, writeErr.HasErrorCode(11000))

	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
	require.NoError(t, err)
	require.NoError(t, cursor.All(ctx, &docs))
	assert.Equal(t, []any{valid1, valid2}, docs)
}
