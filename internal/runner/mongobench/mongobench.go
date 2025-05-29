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
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/runner"
)

// benchmark contains numerical results of a benchmark operation.
//
// The mongodb-benchmarking test produces measurements each second while the benchmark is running,
// and each index of the slices corresponds to a measurement from each second.
type benchmark struct {
	ts        []int64   // timestamp (epoch seconds)
	counts    []int64   // total document count
	meanRates []float64 // mean operation rate in docs/sec
	m1Rates   []float64 // moving average rates over 1 minute
	m5Rates   []float64 // moving average rates over 5 minutes
	m15Rates  []float64 // moving average rates over 15 minutes
}

// Values implements [config.Measurements] interface.
func (m *benchmark) Values() any {
	return map[string]any{
		"ts":         m.ts,
		"counts":     m.counts,
		"mean_rates": m.meanRates,
		"m1_rates":   m.m1Rates,
		"m5_rates":   m.m5Rates,
		"m15_rates":  m.m15Rates,
	}
}

// mongoBench represents `mongoBench` runner.
type mongoBench struct {
	p *config.RunnerParamsMongoBench
	l *slog.Logger
}

// New creates a new runner with given parameters.
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
func parseMeasurements(r *bufio.Reader) (*benchmark, error) {
	// cannot use [csv.NewReader] because the file does not contain valid CSV,
	// it contains 7 header fields while record lines contain 6 fields,
	// so we parse it manually and assume the last field `mean_rate` is missing
	if _, _, err := r.ReadLine(); err != nil {
		return nil, err
	}

	var ts, counts []int64
	var meanRates, m1Rates, m5Rates, m15Rates []float64

	for {
		line, _, err := r.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		record := strings.Split(string(line), ",")

		if len(record) < 6 {
			return nil, errors.New("malformed line: insufficient fields")
		}
		t, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, err
		}

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

		ts = append(ts, t)
		counts = append(counts, count)
		meanRates = append(meanRates, mean)
		m1Rates = append(m1Rates, m1Rate)
		m5Rates = append(m5Rates, m5Rate)
		m15Rates = append(m15Rates, m15Rate)
	}

	return &benchmark{
		ts:        ts,
		counts:    counts,
		meanRates: meanRates,
		m1Rates:   m1Rates,
		m5Rates:   m5Rates,
		m15Rates:  m15Rates,
	}, nil
}

// readMeasurements reads the measurements from the file with the given name.
func readMeasurements(fileName string) (*benchmark, error) {
	relPath := filepath.Join("..", "projects", fileName)

	f, err := os.Open(relPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = f.Close()
	}()

	return parseMeasurements(bufio.NewReader(f))
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

	res := make(map[string]config.TestResult)

	for _, fileName := range fileNames {
		var m *benchmark

		if m, err = readMeasurements(fileName); err != nil {
			return nil, err
		}

		op := strings.TrimSuffix(strings.TrimPrefix(fileName, "benchmark_results_"), ".csv")
		res[op] = config.TestResult{
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

// check interface.
var _ config.Measurements = (*benchmark)(nil)
