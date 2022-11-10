package diff

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestFreeMonitoring(t *testing.T) {
	t.Parallel()

	//ctx, db := setupWithOpts(t, &setupOpts{DatabaseName: "admin"})
	ctx, db := setup(t)
	db = db.Client().Database("admin")

	for name, tc := range map[string]struct {
		setCommand      *bson.D
		setResponse     *bson.D
		setErrorMongoDB *mongo.CommandError

		getCommand          *bson.D
		getResponseFerretDB *bson.D
		// each key is an equivalent of the key in the response bson.D
		// every field will be compared (same as the type).
		// strings inside the map are handled as regex expressions.
		getResponseMongoDB map[string]interface{}
		getErrorMongoDB    *mongo.CommandError
	}{
		"Enable": {
			setCommand:  &bson.D{{"setFreeMonitoring", 1}, {"state", "enable"}},
			setResponse: &bson.D{},

			getCommand:          &bson.D{{"getFreeMonitoringStatus", 1}},
			getResponseFerretDB: &bson.D{{"state", "undecided"}, {"message", "monitoring is undecided"}, {"ok", float64(1)}},
			getResponseMongoDB: map[string]interface{}{
				"state": "undecided",
				"message": "To see your monitoring data, navigate to the unique URL below. Anyone you share the URL with " +
					"will also be able to view this page. You can disable monitoring at any time by running db.disableFreeMonitoring().",
				"ok":           float64(1),
				"url":          "https://cloud.mongodb.com/freemonitoring/cluster/\\w+",
				"userReminder": "Free Monitoring URL:\nhttps://cloud.mongodb.com/freemonitoring/cluster/\\w+",
			},

			//GetResponseMongoDB: &
			//	{"state", "disabled"},
			//	{"message", "To see your monitoring data, navigate to the unique URL below. Anyone you share the URL with " +
			//		"will also be able to view this page. You can disable monitoring at any time by running db.disableFreeMonitoring()."},
			//	{"ok", float64(1)},
			//	{"url", "https://cloud.mongodb.com/freemonitoring/cluster/HFRVTS4MFL2AWMVIP6ICTYL4FRVSKMWE"},
			//	{"userReminder", "Free Monitoring URL:\nhttps://cloud.mongodb.com/freemonitoring/cluster/HFRVTS4MFL2AWMVIP6ICTYL4FRVSKMWE"},
			//},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {

			t.Run("MongoDB", func(t *testing.T) {

				if tc.setCommand != nil {
					var actual bson.D
					err := db.RunCommand(ctx, tc.setCommand).Decode(&actual)

					if tc.setErrorMongoDB != nil {
						AssertEqualError(t, *tc.setErrorMongoDB, err)
						return
					}
					assert.NoError(t, err)

				}

				if tc.getCommand != nil {
					var actual bson.D
					err := db.RunCommand(ctx, tc.getCommand).Decode(&actual)

					if tc.getErrorMongoDB != nil {
						AssertEqualError(t, *tc.getErrorMongoDB, err)
						return
					}
					assert.NoError(t, err)

					actualMap := actual.Map()
					// TODO: check if getResponseMongoDB is set
					// TODO: check if keys are the same

					for key, actualVal := range actualMap {
						expectedVal := tc.getResponseMongoDB[key]

						// handle strings as regex

						switch actualVal := actualVal.(type) {
						case string:
							expectedVal, ok := expectedVal.(string)
							if !ok {
								assert.Equal(t, expectedVal, actualVal)
							}

							assert.Regexp(t, regexp.MustCompile(expectedVal), actualVal)

						default:
							assert.Equal(t, expectedVal, actualVal)

						}

						// handle other types normally
					}
				}
			})
		})
	}

}
