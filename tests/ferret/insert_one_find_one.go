package ferret

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var TestInsertOneFindOne = func(ctx context.Context, db *mongo.Database, t *testing.T) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		collection := db.Collection(collectionName(t))

		for name, v := range map[string]any{
			"double":    math.MaxFloat64,
			"string":    "string",
			"document":  bson.D{{"foo", "bar"}},
			"array":     bson.A{"baz", "qux"},
			"binary":    primitive.Binary{Subtype: 0xFF, Data: []byte{0x01, 0x02, 0x03}},
			"false":     false,
			"true":      true,
			"datetime":  primitive.DateTime(time.Now().UnixMilli()),
			"null":      nil,
			"regex":     primitive.Regex{Pattern: "pattern", Options: "i"},
			"int32":     int32(math.MaxInt32),
			"timestamp": primitive.Timestamp{T: 1, I: 2},
			"int64":     int64(math.MaxInt64),
		} {
			name, v := name, v
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				id := primitive.NewObjectID()
				doc := bson.D{{"_id", id}, {"value", v}}
				_, err := collection.InsertOne(ctx, doc)
				require.NoError(t, err)

				var actual any
				err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actual)
				require.NoError(t, err)
				assert.Equal(t, doc, actual)
			})
		}
	}
}
