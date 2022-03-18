package ferret

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var TestInsertMany = func(ctx context.Context, db *mongo.Database, t *testing.T) func(t *testing.T) {
	return func(t *testing.T) {
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
	}
}
