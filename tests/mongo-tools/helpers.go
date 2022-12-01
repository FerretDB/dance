package mongotools

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// execCommand runs command on docker mongosh container.
func execCommand(command string, args ...string) error {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	dockerArgs := append([]string{"compose", "exec", "mongosh", command}, args...)
	cmd := exec.Command(bin, dockerArgs...)
	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(dockerArgs, " "), err)
	}

	return nil
}

// getDatabaseState gets all documents from each collection in provided database and puts
// them in a map keyed with collection names
func getDatabaseState(t *testing.T, ctx context.Context, db *mongo.Database) map[string][]bson.D {
	dbState := make(map[string][]bson.D)

	doc := struct {
		Cursor struct {
			FirstBatch []struct {
				Name string `bson:"name"`
			} `bson:"firstBatch"`
		} `bson:"cursor"`
	}{}

	err := db.RunCommand(ctx, bson.D{{"listCollections", 1}}).Decode(&doc)
	require.NoError(t, err)

	var collections []string
	for _, batch := range doc.Cursor.FirstBatch {
		collections = append(collections, batch.Name)
	}

	for _, coll := range collections {
		cur, err := db.Collection(coll).Find(ctx, bson.D{{}})
		require.NoError(t, err)

		var res []bson.D
		require.NoError(t, cur.All(ctx, &res))

		dbState[coll] = res
	}

	return dbState
}
