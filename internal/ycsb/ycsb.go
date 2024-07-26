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

// Package ycsb contains `ycsb` runner.
package ycsb

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/FerretDB/dance/internal/config"
)

// Measurements stores go-ycsb results.
type Measurements struct {
	Takes    time.Duration
	Count    int64
	OPS      float64
	Avg      time.Duration
	Min      time.Duration
	Max      time.Duration
	Perc50   time.Duration
	Perc90   time.Duration
	Perc95   time.Duration
	Perc99   time.Duration
	Perc999  time.Duration
	Perc9999 time.Duration
}

// Run runs `go-ycsb`.
//
// It loads and runs a YCSB workload.
// Properties defined in the YAML file will override properties defined in the workload parameter file.
func Run(ctx context.Context, dir string, args []string) (*config.TestResults, error) {
	bin := filepath.Join("..", "bin", "go-ycsb")
	if _, err := os.Stat(bin); err != nil {
		return nil, err
	}

	// because we set cmd.Dir, the relative path here is different
	bin = filepath.Join("..", bin)

	// load workload

	cliArgs := []string{"load", "mongodb", "-P", args[0]}
	for _, p := range args[1:] {
		cliArgs = append(cliArgs, "-p", p)
	}

	cmd := exec.CommandContext(ctx, bin, cliArgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	// FIXME ?
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// run workload with almost the same args

	cliArgs[0] = "run"

	cmd = exec.CommandContext(ctx, bin, cliArgs...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	defer pipe.Close()

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	var buf strings.Builder
	mw := io.MultiWriter(os.Stdout, &buf)

	go io.Copy(mw, pipe)

	res := &config.TestResults{
		TestResults: map[string]config.TestResult{
			dir: {
				Status: config.Pass,
			},
		},
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()

	switch err {
	case nil:
		fmt.Printf("Parsed metrics: %+v\n\n", parseMeasurements(buf.String()))
	default:
		res.TestResults[dir] = config.TestResult{
			Status: config.Fail,
			Output: err.Error(),
		}
	}

	return res, nil
}

// parseMeasurements parses go-ycsb results.
func parseMeasurements(output string) map[string]Measurements {
	res := make(map[string]Measurements)

	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		prefix := fields[0]

		switch prefix {
		case "TOTAL", "READ", "INSERT", "UPDATE":
			var parsedPrefix string
			var takes, ops float64
			var count, avg, vmin, vmax, perc50, perc90, perc95, perc99, perc999, perc9999 int64

			// It is enough to use fmt.Sscanf for parsing the data as the string produced by go-ycsb has fixed format:
			// https://github.com/pingcap/go-ycsb/blob/fe11c4783b57703465ec7d36fcc4268979001d1a/pkg/measurement/measurement.go#L28
			_, err := fmt.Sscanf(line,
				"%s - Takes(s): %f, Count: %d, OPS: %f, Avg(us): %d, Min(us): %d, Max(us): %d, "+
					"50th(us): %d, 90th(us): %d, 95th(us): %d, 99th(us): %d, 99.9th(us): %d, 99.99th(us): %d",
				&parsedPrefix, &takes, &count, &ops, &avg, &vmin, &vmax, &perc50, &perc90, &perc95, &perc99, &perc999, &perc9999,
			)
			if err != nil {
				panic(err)
			}

			res[prefix] = Measurements{
				Takes:    time.Duration(takes * float64(time.Second)),
				Count:    count,
				OPS:      ops,
				Avg:      time.Duration(avg * int64(time.Microsecond)),
				Min:      time.Duration(vmin * int64(time.Microsecond)),
				Max:      time.Duration(vmax * int64(time.Microsecond)),
				Perc50:   time.Duration(perc50 * int64(time.Microsecond)),
				Perc90:   time.Duration(perc90 * int64(time.Microsecond)),
				Perc95:   time.Duration(perc95 * int64(time.Microsecond)),
				Perc99:   time.Duration(perc99 * int64(time.Microsecond)),
				Perc999:  time.Duration(perc999 * int64(time.Microsecond)),
				Perc9999: time.Duration(perc9999 * int64(time.Microsecond)),
			}
		default:
			// string doesn't contain metrics, do nothing
		}
	}

	return res
}
