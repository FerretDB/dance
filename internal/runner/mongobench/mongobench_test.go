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
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFileNames(t *testing.T) {
	t.Parallel()

	//nolint:lll // verbatim output
	output := strings.NewReader(`

2025/05/26 15:28:13 Collection dropped. Starting new rate test...
2025/05/26 15:28:14 Timestamp: 1748240894, Document Count: 20138, Mean Rate: 20134.59 docs/sec, m1_rate: 0.00, m5_rate: 0.00, m15_rate: 0.00
Benchmarking completed. Results saved to benchmark_results_insert.csv
2025/05/26 15:28:15 Starting update test...
2025/05/26 15:28:16 Timestamp: 1748240896, Document Count: 18717, Mean Rate: 18716.72 docs/sec, m1_rate: 0.00, m5_rate: 0.00, m15_rate: 0.00
Benchmarking completed. Results saved to benchmark_results_update.csv
2025/05/26 15:28:16 Starting delete test...
2025/05/26 15:28:17 Timestamp: 1748240897, Document Count: 20690, Mean Rate: 20689.05 docs/sec, m1_rate: 0.00, m5_rate: 0.00, m15_rate: 0.00
Benchmarking completed. Results saved to benchmark_results_delete.csv
2025/05/26 15:28:18 Collection dropped. Starting new rate test...
2025/05/26 15:28:19 Timestamp: 1748240899, Document Count: 13524, Mean Rate: 13522.22 docs/sec, m1_rate: 756.80, m5_rate: 756.80, m15_rate: 756.80
2025/05/26 15:28:20 Timestamp: 1748240900, Document Count: 21505, Mean Rate: 10748.08 docs/sec, m1_rate: 756.80, m5_rate: 756.80, m15_rate: 756.80
2025/05/26 15:28:21 Timestamp: 1748240901, Document Count: 27081, Mean Rate: 9027.36 docs/sec, m1_rate: 756.80, m5_rate: 756.80, m15_rate: 756.80
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

func TestReadResult(t *testing.T) {
	t.Parallel()

	results := `t,count,mean,m1_rate,m5_rate,m15_rate,mean_rate
1748240899,13524,13522.216068,756.800000,756.800000,756.800000
1748240900,21505,10748.078321,756.800000,756.800000,756.800000
1748240901,27081,9027.363746,756.800000,756.800000,756.800000
1748240902,30000,8288.694528,756.800000,756.800000,756.800000
`
	f, err := os.CreateTemp(t.TempDir(), "benchmark_results_delete")
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, f.Close())
	})

	_, err = f.Write([]byte(results))
	require.NoError(t, err)

	actual, err := readResult(f.Name())
	require.NoError(t, err)

	expected := map[string]float64{
		"count":    30000,
		"mean":     8288.694528,
		"m1_rate":  756.800000,
		"m5_rate":  756.800000,
		"m15_rate": 756.800000,
	}
	assert.Equal(t, expected, actual)
}
