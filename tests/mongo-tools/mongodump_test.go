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

	"github.com/FerretDB/dance/tests/common"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/stretchr/testify/require"
)

func TestMongodump(t *testing.T) {
	// TODO ensure FerretDB's `task run` and ferretdb_mongodb compatibility
	ctx, db := common.Setup(t)
	var err error

	_, err = db.Collection("mongodump").InsertOne(ctx, bson.D{{"foo", "bar"}})
	require.NoError(t, err)

	// TODO: collections are not listed
	//cur, err := db.ListCollections(ctx, bson.D{{}})
	//require.NoError(t, err)

	//var res bson.D
	//cur.Decode(&res)

	assert.Equal(t, bson.D{}, res)

	buffer := bytes.NewBuffer([]byte{})
	err = runCommand("docker", []string{"compose", "exec", "mongosh", "mongodump",
		"mongodb://dance_ferretdb:27017/testmongodump",
		"--verbose",
	}, buffer)
	require.NoError(t, err)

	// We can remove this and just check if changes are applied
	//out := buffer.String()
	//t.Log(out)
	//assert.Equal(t, "dumping up to 1 collections in parallel\n", strings.Split(out, "\t")[1])
	buffer.Reset()

	err = runCommand("docker", []string{"compose", "exec", "mongosh", "mongorestore",
		"mongodb://dance_ferretdb:27017/testmongodump",
		"--verbose",
	}, buffer)
	require.NoError(t, err)
}
