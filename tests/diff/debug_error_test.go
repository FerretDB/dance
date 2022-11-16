package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDebugError(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)

	for name, tc := range map[string]struct {
		input       string
		expectedErr error
		expectedRes bson.D
	}{
		"ok": {
			input: "ok",
		},
		"panic": {
			input: "panic",
		},
		"codeZero": {
			input: "0",
		},
		"codeOne": {
			input: "1",
		},
		"codeNotLabeled": {
			input: "33333",
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var actual bson.D
			err := db.RunCommand(ctx, bson.D{{"debugError", tc.input}}).Decode(&actual)

			t.Run("FerretDB", func(t *testing.T) {
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr, err)
					return
				}
				assert.NoError(t, err)

				assert.Equal(t, tc.expectedRes, actual)

			})
		})
	}

}
