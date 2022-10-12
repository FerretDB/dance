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

// TestDocumentValidation tests validation rules for documents insert and update.
func TestDocumentValidation(t *testing.T) {
	t.Parallel()
	ctx, db := setup(t)

	// initiate a collection with a valid document, so we have something to update
	collection := db.Collection("document-validation")
	_, err := collection.InsertOne(ctx, bson.D{
		{"_id", "valid"},
		{"v", int32(42)},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		expectedErrMongoDB  *mongo.CommandError
		expectedErrFerretDB *mongo.CommandError
		doc                 bson.D
	}{
		"KeyIsNotUTF8": {
			doc:                bson.D{{"\xF4\x90\x80\x80", int32(12)}}, //  the key is out of range for UTF-8
			expectedErrMongoDB: nil,
			expectedErrFerretDB: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Invalid document, reason: invalid key: "\xf4\x90\x80\x80" (not a valid UTF-8 string).`,
			},
		},
		"KeyContains$": {
			doc:                bson.D{{"v$", int32(12)}},
			expectedErrMongoDB: nil,
			expectedErrFerretDB: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Invalid document, reason: invalid key: "v$" (key mustn't contain $).`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Run("FerretDB", func(t *testing.T) {
				_, err := collection.InsertOne(ctx, tc.doc)
				if tc.expectedErrFerretDB != nil {
					require.NotNil(t, err)
					AssertEqualError(t, *tc.expectedErrFerretDB, err)
				} else {
					require.Nil(t, err)
				}

				_, err = collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", tc.doc}})
				if tc.expectedErrFerretDB != nil {
					require.NotNil(t, err)
					AssertEqualError(t, *tc.expectedErrFerretDB, err)
				} else {
					require.Nil(t, err)
				}
			})

			t.Run("MongoDB", func(t *testing.T) {
				_, err := collection.InsertOne(ctx, tc.doc)
				if tc.expectedErrMongoDB != nil {
					require.NotNil(t, err)
					AssertEqualError(t, *tc.expectedErrMongoDB, err)
				} else {
					require.Nil(t, err)
				}

				_, err = collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", tc.doc}})
				if tc.expectedErrMongoDB != nil {
					require.NotNil(t, err)
					AssertEqualError(t, *tc.expectedErrMongoDB, err)
				} else {
					require.Nil(t, err)
				}
			})
		})
	}
}
