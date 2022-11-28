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

package mongotools

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMongodump(t *testing.T) {
	// TODO ensure FerretDB's `task run` and ferretdb_mongodb compatibility

	buffer := bytes.NewBuffer([]byte{})
	err := runCommand("docker", []string{"compose", "exec", "mongosh", "mongodump",
		"mongodb://127.0.0.1:27017/test",
		"--verbose",
	}, buffer)
	assert.NoError(t, err)
	t.Log(buffer.String())
}
