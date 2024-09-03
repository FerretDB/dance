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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMeasurements(t *testing.T) {
	t.Parallel()

	//nolint:lll // verbatim output
	output := strings.NewReader(`
Using request distribution 'uniform' a keyrange of [0 999]
Connected to MongoDB!
***************** properties *****************
"workload"="core"
"recordcount"="1000"
"outputstyle"="json"
"readproportion"="0.5"
"scanproportion"="0"
"requestdistribution"="uniform"
"dotransactions"="true"
"mongodb.url"="mongodb://127.0.0.1:37001/"
"readallfields"="true"
"updateproportion"="0.5"
"insertproportion"="0"
"operationcount"="50000"
"command"="run"
**********************************************
[{"50th(us)":"287","90th(us)":"328","95th(us)":"350","99.99th(us)":"1312","99.9th(us)":"777","99th(us)":"431","Avg(us)":"292","Count":"17129","Max(us)":"2407","Min(us)":"175","OPS":"1713.1","Operation":"READ","Takes(s)":"10.0"},{"50th(us)":"284","90th(us)":"325","95th(us)":"346","99.99th(us)":"2059","99.9th(us)":"781","99th(us)":"419","Avg(us)":"289","Count":"34279","Max(us)":"2407","Min(us)":"171","OPS":"3428.1","Operation":"TOTAL","Takes(s)":"10.0"},{"50th(us)":"281","90th(us)":"321","95th(us)":"342","99.99th(us)":"2059","99.9th(us)":"794","99th(us)":"410","Avg(us)":"286","Count":"17150","Max(us)":"2217","Min(us)":"171","OPS":"1715.1","Operation":"UPDATE","Takes(s)":"10.0"}]
**********************************************
Run finished, takes 14.676689042s
[{"50th(us)":"288","90th(us)":"331","95th(us)":"353","99.99th(us)":"1312","99.9th(us)":"757","99th(us)":"427","Avg(us)":"293","Count":"25030","Max(us)":"2407","Min(us)":"175","OPS":"1705.6","Operation":"READ","Takes(s)":"14.7"},{"50th(us)":"285","90th(us)":"328","95th(us)":"349","99.99th(us)":"1913","99.9th(us)":"777","99th(us)":"421","Avg(us)":"291","Count":"50000","Max(us)":"2407","Min(us)":"171","OPS":"3406.9","Operation":"TOTAL","Takes(s)":"14.7"},{"50th(us)":"282","90th(us)":"324","95th(us)":"345","99.99th(us)":"2151","99.9th(us)":"786","99th(us)":"411","Avg(us)":"288","Count":"24970","Max(us)":"2261","Min(us)":"171","OPS":"1701.4","Operation":"UPDATE","Takes(s)":"14.7"}]
`)

	actual, err := parseOutput(output)
	require.NoError(t, err)

	expected := map[string]map[string]float64{
		"read": {
			"takes":    14.7,
			"count":    25030,
			"ops":      1705.6,
			"avg":      0.000293,
			"min":      0.000175,
			"max":      0.002407,
			"perc50":   0.000288,
			"perc90":   0.000331,
			"perc95":   0.000353,
			"perc99":   0.000427,
			"perc999":  0.000757,
			"perc9999": 0.001312,
		},
		"update": {
			"takes":    14.7,
			"count":    24970,
			"ops":      1701.4,
			"avg":      0.000288,
			"min":      0.000171,
			"max":      0.002261,
			"perc50":   0.000282,
			"perc90":   0.000324,
			"perc95":   0.000345,
			"perc99":   0.000411,
			"perc999":  0.000786,
			"perc9999": 0.002151,
		},
		"total": {
			"takes":    14.7,
			"count":    50000,
			"ops":      3406.9,
			"avg":      0.000291,
			"min":      0.000171,
			"max":      0.002407,
			"perc50":   0.000285,
			"perc90":   0.000328,
			"perc95":   0.000349,
			"perc99":   0.000421,
			"perc999":  0.000777,
			"perc9999": 0.001913,
		},
	}

	assert.Equal(t, expected, actual)
}
