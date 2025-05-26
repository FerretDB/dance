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

// Package mongobench provides `mongobench` runner.
package mongobench

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/runner"
)

// mongoBench represents `mongoBench` runner.
type mongoBench struct {
	p *config.RunnerParamsMongoBench
	l *slog.Logger
}

// New creates a new `mongoBench` runner with given parameters.
func New(params *config.RunnerParamsMongoBench, l *slog.Logger) (runner.Runner, error) {
	return &mongoBench{
		p: params,
		l: l,
	}, nil
}

// parseFileNames parses the file names that store benchmark results.
// Each operation is stored in different files such as `benchmark_results_insert.csv`,
// `benchmark_results_update.csv`, `benchmark_results_delete.csv` and `benchmark_results_upsert.csv`.
func parseFileNames(r io.Reader) ([]string, error) {
	var fileNames []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "Benchmarking completed. Results saved to ") {
			continue
		}

		// parse file name from the line such as `Benchmarking completed. Results saved to benchmark_results_delete.csv`
		fileName := strings.TrimSpace(strings.TrimPrefix(line, "Benchmarking completed. Results saved to "))
		fileNames = append(fileNames, fileName)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(fileNames) == 0 {
		return nil, errors.New("no benchmark result files found")
	}

	return fileNames, nil
}

// parseMeasurements reads the mongo-bench results from the reader.
func parseMeasurements(op string, r *bufio.Reader) (map[string]map[string]float64, error) {
	res := make(map[string]map[string]float64)

	// cannot use [csv.NewReader] because the file does not contain valid CSV,
	// it contains 7 header fields while record lines contain 6 fields,
	// so we parse it manually and assume last field `mean_rate` is missing
	//
	//t,count,mean,m1_rate,m5_rate,m15_rate,mean_rate
	//1748240899,13524,13522.216068,756.800000,756.800000,756.800000

	// ignore header
	if _, _, err := r.ReadLine(); err != nil {
		return nil, err
	}

	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		record := strings.Split(string(line), ",")

		i, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, err
		}

		t := time.Unix(i, 0)

		count, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return nil, err
		}

		mean, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, err
		}

		m1Rate, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, err
		}

		m5Rate, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return nil, err
		}

		m15Rate, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			return nil, err
		}

		// FIXME each record is a measurement produced each second during the benchmark is running,
		// not a single measurement of an operation
		res[fmt.Sprintf("%s_%d", op, t.Unix())] = map[string]float64{
			"t":        float64(t.Unix()), // timestamp (epoch seconds)
			"count":    float64(count),    // total document count
			"mean":     mean,              // mean operation rate in docs/sec
			"m1_rate":  m1Rate,            // moving average rates over 1 minute
			"m5_rate":  m5Rate,            // moving average rates over 5 minutes
			"m15_rate": m15Rate,           // moving average rates over 15 minutes
		}
	}

	return res, nil
}

// run runs given command in the given directory and returns parsed results.
func run(ctx context.Context, args []string, dir string) (map[string]config.TestResult, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	defer pipe.Close()

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	fileNames, err := parseFileNames(io.TeeReader(pipe, os.Stdout))
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	ms := make(map[string]map[string]float64)

	for _, fileName := range fileNames {
		relPath := filepath.Join("..", fileName)

		var f *os.File

		if f, err = os.Open(relPath); err != nil {
			return nil, err
		}

		defer f.Close()

		op := strings.TrimSuffix(strings.TrimPrefix(fileName, "benchmark_results_"), ".csv")

		var m map[string]map[string]float64

		if m, err = parseMeasurements(op, bufio.NewReader(f)); err != nil {
			return nil, err
		}

		for k, v := range m {
			ms[k] = v
		}
	}

	res := make(map[string]config.TestResult)
	for t, m := range ms {
		res[t] = config.TestResult{
			Status:       config.Pass,
			Measurements: m,
		}
	}

	return res, nil
}

// Run implements [runner.Runner] interface.
func (y *mongoBench) Run(ctx context.Context) (map[string]config.TestResult, error) {
	bin := filepath.Join("..", "bin", "mongodb-benchmarking")
	if _, err := os.Stat(bin); err != nil {
		return nil, err
	}

	bin, err := filepath.Abs(bin)
	if err != nil {
		return nil, err
	}

	args := append([]string{bin}, y.p.Args...)

	y.l.InfoContext(ctx, "Run", slog.String("cmd", strings.Join(args, " ")))

	return run(ctx, args, y.p.Dir)
}
