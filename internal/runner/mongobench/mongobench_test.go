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

package mongobench

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFileNames(t *testing.T) {
	t.Parallel()

	//nolint:lll // verbatim output
	output := strings.NewReader(`

2025/05/26 11:27:41 Collection dropped. Starting new rate test...
2025/05/26 11:27:42 Timestamp: 1748226462, Document Count: 19776, Mean Rate: 19773.41 docs/sec, m1_rate: 0.00, m5_rate: 0.00, m15_rate: 0.00
Benchmarking completed. Results saved to benchmark_results_insert.csv
2025/05/26 11:27:42 Starting update test...
2025/05/26 11:27:43 Timestamp: 1748226463, Document Count: 17893, Mean Rate: 17888.70 docs/sec, m1_rate: 0.00, m5_rate: 0.00, m15_rate: 0.00
Benchmarking completed. Results saved to benchmark_results_update.csv
2025/05/26 11:27:43 Starting delete test...
Benchmarking completed. Results saved to benchmark_results_delete.csv
2025/05/26 11:27:44 Collection dropped. Starting new rate test...
2025/05/26 11:27:45 Timestamp: 1748226465, Document Count: 16680, Mean Rate: 16677.66 docs/sec, m1_rate: 0.00, m5_rate: 0.00, m15_rate: 0.00
Benchmarking completed. Results saved to benchmark_results_upsert.csv
`)

	actual, err := parseFileNames(output)
	require.NoError(t, err)

	expected := []string{
		"benchmark_results_insert.csv",
		"benchmark_results_update.csv",
		"benchmark_results_delete.csv",
		"benchmark_results_upsert.csv",
	}

	assert.Equal(t, expected, actual)
}

func TestParseMeasurements(t *testing.T) {
	t.Parallel()

	//nolint:lll // verbatim results
	results := strings.NewReader(
		`t,count,mean,m1_rate,m5_rate,m15_rate,mean_rate
1748240899,13524,13522.216068,756.800000,756.800000,756.800000
1748240900,21505,10748.078321,756.800000,756.800000,756.800000
1748240901,27081,9027.363746,756.800000,756.800000,756.800000
1748240902,30000,8288.694528,756.800000,756.800000,756.800000`,
	)

	actual, err := parseMeasurements("upsert", bufio.NewReader(results))
	require.NoError(t, err)

	expected := map[string]map[string]float64{
		"upsert_1748240899": {
			"t":        1748240899,
			"count":    13524,
			"mean":     13522.216068,
			"m1_rate":  756.800000,
			"m5_rate":  756.800000,
			"m15_rate": 756.800000,
		},
		"upsert_1748240900": {
			"t":        1748240900,
			"count":    21505,
			"mean":     10748.078321,
			"m1_rate":  756.800000,
			"m5_rate":  756.800000,
			"m15_rate": 756.800000,
		},
		"upsert_1748240901": {
			"t":        1748240901,
			"count":    27081,
			"mean":     9027.363746,
			"m1_rate":  756.800000,
			"m5_rate":  756.800000,
			"m15_rate": 756.800000,
		},
		"upsert_1748240902": {
			"t":        1748240902,
			"count":    30000,
			"mean":     8288.694528,
			"m1_rate":  756.800000,
			"m5_rate":  756.800000,
			"m15_rate": 756.800000,
		},
	}

	assert.Equal(t, expected, actual)
}
