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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/dance/tests/common"
)

func TestErrorMessages(t *testing.T) {
	t.Parallel()

	ctx, db := common.Setup(t)
	db = db.Client().Database("admin")

	var doc bson.D
	err := db.RunCommand(ctx, bson.D{{"getParameter", bson.D{{"allParameters", "1"}}}}).Decode(&doc)
	var actual mongo.CommandError
	require.ErrorAs(t, err, &actual)
	assert.Equal(t, int32(14), actual.Code)
	assert.Equal(t, "TypeMismatch", actual.Name)

	t.Run("FerretDB", func(t *testing.T) {
		// our error message is better
		expected := mongo.CommandError{
			Code:    14,
			Name:    "TypeMismatch",
			Message: "BSON field 'allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
		}
		AssertEqualError(t, expected, actual)
	})

	t.Run("MongoDB", func(t *testing.T) {
		// closing single quote is in the wrong place
		expected := mongo.CommandError{
			Code: 14,
			Name: "TypeMismatch",
			Message: "BSON field 'getParameter.allParameters' is the wrong type 'string', " +
				"expected types '[bool, long, int, decimal, double']",
		}
		AssertEqualError(t, expected, actual)
	})
}
