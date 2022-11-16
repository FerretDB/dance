package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestDebugError(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

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
			//t.Parallel()
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
		})
	}

}
