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
	"go.mongodb.org/mongo-driver/mongo"
)

func TestDebugError(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		input       string
		expectedErr *mongo.CommandError
		expectedRes bson.D
	}{
		"ok": {
			input:       "ok",
			expectedRes: bson.D{{"ok", float64(1)}},
		},
		"codeZero": {
			input: "0",
			expectedErr: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "Unset",
			},
		},
		"codeOne": {
			input: "1",
			expectedErr: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "InternalError",
			},
		},
		"codeNotLabeled": {
			input: "33333",
			expectedErr: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "ErrorCode(33333)",
			},
		},
		// TODO: https://github.com/FerretDB/dance/issues/249
		//"panic": {
		//	input: "panic",
		//	expectedErr: &mongo.CommandError{
		//		Code:    0,
		//		Message: "connection(127.0.0.1:27017[-22]) socket was unexpectedly closed: EOF",
		//		Labels:  []string{"NetworkError"},
		//	},
		//},
		"string": {
			input: "foo",
			expectedErr: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "foo",
			},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, db := setup(t)
			var actual bson.D
			err := db.RunCommand(ctx, bson.D{{"debugError", tc.input}}).Decode(&actual)

			t.Run("FerretDB", func(t *testing.T) {
				if tc.expectedErr != nil {
					AssertEqualError(t, *tc.expectedErr, err)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRes, actual)
			})

			t.Run("MongoDB", func(t *testing.T) {
				expectedErr := mongo.CommandError{
					Code:    59,
					Message: "no such command: 'debugError'",
					Name:    "CommandNotFound",
				}

				AssertEqualError(t, expectedErr, err)
			})
		})
	}
}
