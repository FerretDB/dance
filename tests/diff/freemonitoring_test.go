package diff

import (
	"testing"

	"golang.org/x/exp/slices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestFreeMonitoring(t *testing.T) {
	t.Parallel()

	ctx, db := setup(t)
	db = db.Client().Database("admin")

	res := db.RunCommand(ctx, bson.D{{"setFreeMonitoring", 1}, {"action", "enable"}})
	require.NoError(t, res.Err())

	var actual bson.D
	err := db.RunCommand(ctx, bson.D{{"getFreeMonitoringStatus", 1}}).Decode(&actual)
	assert.NoError(t, err)

	t.Run("MongoDB", func(t *testing.T) {
		t.Parallel()

		expectedKeys := []string{"message", "url", "userReminder", "ok", "state"}
		var actualKeys []string

		for k, v := range actual.Map() {
			actualKeys = append(actualKeys, k)
			switch k {
			case "state":
				assert.Equal(t, "enabled", v)
			case "message":
				msg := "To see your monitoring data, navigate to the unique URL below. Anyone you share the URL with " +
					"will also be able to view this page. You can disable monitoring at any time by running db.disableFreeMonitoring()."
				assert.Equal(t, msg, v)
			case "ok":
				assert.Equal(t, float64(1), v)
			case "url":
				assert.Regexp(t, "https://cloud.mongodb.com/freemonitoring/cluster/\\w+", v)
			}
		}

		slices.Sort(expectedKeys)
		slices.Sort(actualKeys)

		assert.Equal(t, expectedKeys, actualKeys)
	})

	t.Run("FerretDB", func(t *testing.T) {
		t.Parallel()

		expectedRes := bson.D{{"state", "enabled"}, {"message", "monitoring is enabled"}, {"ok", float64(1)}}
		assert.Equal(t, actual, expectedRes)
	})
}
