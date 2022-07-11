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
	"go.mongodb.org/mongo-driver/bson"
)

func TestNullStrings(t *testing.T) {
	t.Parallel()
	ctx, db := setup(t)

	_, err := db.Collection("insert").InsertOne(ctx, bson.D{
		{"_id", "document"},
		{"a", string([]byte{0})},
	})

	t.Run("FerretDB", func(t *testing.T) {
		require.Error(t, err)
		assert.Regexp(t, "^.* unsupported Unicode escape sequence .*$", err.Error())
	})

	t.Run("MongoDB", func(t *testing.T) {
		assert.Equal(t, err, nil)
	})
}
