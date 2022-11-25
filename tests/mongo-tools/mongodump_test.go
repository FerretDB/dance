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
	"bufio"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestMongodump(t *testing.T) {
	// TODO ensure FerretDB's `task run` and ferretdb_mongodb compatibility
	cli, err := client.NewEnvClient()
	require.NoError(t, err)

	ctx := context.Background()

	conf := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,

		Cmd: []string{"mongodump", "--uri", "mongodb://dance_ferretdb:27017/test"},
	}

	id, err := cli.ContainerExecCreate(ctx, "dance_mongosh", conf)
	require.NoError(t, err)

	resp, err := cli.ContainerExecAttach(ctx, id.ID, types.ExecStartCheck{
		// TODO Tty
	})
	// TODO wait for container
	require.NoError(t, err)

	err = cli.ContainerExecStart(ctx, id.ID, types.ExecStartCheck{})
	require.NoError(t, err)

	scanner := bufio.NewScanner(resp.Reader)

	for scanner.Scan() {
		err = scanner.Err()
		require.NoError(t, err)
		log.Printf(scanner.Text())
	}

}
