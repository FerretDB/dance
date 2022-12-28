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
	"testing"
)

func TestExportImport(t *testing.T) {
	t.Parallel()
}

// mongoimport imports collection from <file> file as <db>/<coll>.
func mongoimport(t *testing.T, file, db, coll string) {
	t.Helper()

	runDockerComposeCommand(
		t,
		"mongoimport",
		"--verbose=2",
		"--db="+db,
		"--collection="+coll,
		"--file="+file,
		"--sort",
		"--drop",
		"--numInsertionWorkers=10",
		"--stopOnError",
		"mongodb://host.docker.internal:27017/",
	)
}
