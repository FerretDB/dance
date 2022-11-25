package diff

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

func setupMongoContainer(t *testing.T) {
	t.Helper()
	// TODO ensure FerretDB's `task run` and ferretdb_mongodb compatibility
	cli, err := client.NewEnvClient()
	require.NoError(err)

	ctx := context.Background()

	conf := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,

		Cmd: []string{"mongodump", "-c", "test"}
	}

	resp, err := cli.ContainerExecAttach(ctx, "dance_mongodb", types.ExecStartCheck{
		// TODO Tty
	})
	// TODO wait for container
	require.NoError(t, err)

}
