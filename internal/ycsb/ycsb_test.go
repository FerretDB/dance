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

package ycsb

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMeasurements(t *testing.T) {
	//nolint:lll // verbatim output
	output := strings.NewReader(strings.TrimSpace(`
Using request distribution 'uniform' a keyrange of [0 4999]
Connected to MongoDB!
***************** properties *****************
"insertproportion"="0"
"requestdistribution"="uniform"
"readproportion"="0.5"
"updateproportion"="0.5"
"command"="run"
"workload"="core"
"dotransactions"="true"
"scanproportion"="0"
"readallfields"="true"
"recordcount"="5000"
"operationcount"="100000"
**********************************************
READ   - Takes(s): 10.0, Count: 19332, OPS: 1933.3, Avg(us): 247, Min(us): 138, Max(us): 1110, 50th(us): 243, 90th(us): 289, 95th(us): 317, 99th(us): 407, 99.9th(us): 661, 99.99th(us): 871
TOTAL  - Takes(s): 10.0, Count: 38456, OPS: 3845.7, Avg(us): 257, Min(us): 138, Max(us): 2773, 50th(us): 252, 90th(us): 302, 95th(us): 331, 99th(us): 441, 99.9th(us): 751, 99.99th(us): 1784
UPDATE - Takes(s): 10.0, Count: 19124, OPS: 1912.6, Avg(us): 268, Min(us): 155, Max(us): 2773, 50th(us): 261, 90th(us): 310, 95th(us): 344, 99th(us): 475, 99.9th(us): 1218, 99.99th(us): 1803
READ   - Takes(s): 20.0, Count: 38557, OPS: 1927.9, Avg(us): 247, Min(us): 138, Max(us): 5299, 50th(us): 243, 90th(us): 285, 95th(us): 307, 99th(us): 390, 99.9th(us): 711, 99.99th(us): 2871
TOTAL  - Takes(s): 20.0, Count: 76910, OPS: 3845.6, Avg(us): 257, Min(us): 138, Max(us): 52799, 50th(us): 251, 90th(us): 297, 95th(us): 320, 99th(us): 417, 99.9th(us): 969, 99.99th(us): 2979
UPDATE - Takes(s): 20.0, Count: 38353, OPS: 1917.8, Avg(us): 268, Min(us): 155, Max(us): 52799, 50th(us): 260, 90th(us): 305, 95th(us): 329, 99th(us): 441, 99.9th(us): 1720, 99.99th(us): 2979
**********************************************
Run finished, takes 25.957181042s
READ   - Takes(s): 26.0, Count: 50216, OPS: 1934.6, Avg(us): 246, Min(us): 138, Max(us): 9735, 50th(us): 243, 90th(us): 284, 95th(us): 305, 99th(us): 381, 99.9th(us): 692, 99.99th(us): 2871
TOTAL  - Takes(s): 26.0, Count: 100000, OPS: 3852.6, Avg(us): 257, Min(us): 138, Max(us): 52799, 50th(us): 251, 90th(us): 296, 95th(us): 318, 99th(us): 405, 99.9th(us): 848, 99.99th(us): 2979
UPDATE - Takes(s): 26.0, Count: 49784, OPS: 1918.0, Avg(us): 267, Min(us): 153, Max(us): 52799, 50th(us): 260, 90th(us): 304, 95th(us): 328, 99th(us): 430, 99.9th(us): 1684, 99.99th(us): 2979
`) + "\n")

	actual, err := parseMeasurements(output)
	require.NoError(t, err)

	expected := map[string]Measurements{
		"READ": {
			Takes:    26 * time.Second,
			Count:    50216,
			OPS:      1934.6,
			Avg:      246 * time.Microsecond,
			Min:      138 * time.Microsecond,
			Max:      9735 * time.Microsecond,
			Perc50:   243 * time.Microsecond,
			Perc90:   284 * time.Microsecond,
			Perc95:   305 * time.Microsecond,
			Perc99:   381 * time.Microsecond,
			Perc999:  692 * time.Microsecond,
			Perc9999: 2871 * time.Microsecond,
		},
		"TOTAL": {
			Takes:    26 * time.Second,
			Count:    100000,
			OPS:      3852.6,
			Avg:      257 * time.Microsecond,
			Min:      138 * time.Microsecond,
			Max:      52799 * time.Microsecond,
			Perc50:   251 * time.Microsecond,
			Perc90:   296 * time.Microsecond,
			Perc95:   318 * time.Microsecond,
			Perc99:   405 * time.Microsecond,
			Perc999:  848 * time.Microsecond,
			Perc9999: 2979 * time.Microsecond,
		},
		"UPDATE": {
			Takes:    26 * time.Second,
			Count:    49784,
			OPS:      1918.0,
			Avg:      267 * time.Microsecond,
			Min:      153 * time.Microsecond,
			Max:      52799 * time.Microsecond,
			Perc50:   260 * time.Microsecond,
			Perc90:   304 * time.Microsecond,
			Perc95:   328 * time.Microsecond,
			Perc99:   430 * time.Microsecond,
			Perc999:  1684 * time.Microsecond,
			Perc9999: 2979 * time.Microsecond,
		},
	}
	assert.Equal(t, expected, actual)
}
