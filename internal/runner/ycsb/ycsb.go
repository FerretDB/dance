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
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/runner"
)

// measurement represents a single object in go-ycsb JSON array output.
type measurement struct {
	Operation  string  `json:"Operation"`
	TakesS     float64 `json:"Takes(s),string"`
	Count      int     `json:"Count,string"`
	OPS        float64 `json:"OPS,string"`
	AvgUs      float64 `json:"Avg(us),string"`
	MinUs      float64 `json:"Min(us),string"`
	MaxUs      float64 `json:"Max(us),string"`
	Perc50Us   float64 `json:"50th(us),string"`
	Perc90Us   float64 `json:"90th(us),string"`
	Perc95Us   float64 `json:"95th(us),string"`
	Perc99Us   float64 `json:"99th(us),string"`
	Perc999Us  float64 `json:"99.9th(us),string"`
	Perc9999Us float64 `json:"99.99th(us),string"`
}

// ycsb represents `ycsb` runner.
type ycsb struct {
	p *config.RunnerParamsYCSB
	l *slog.Logger
}

// New creates a new `ycsb` runner with given parameters.
func New(params *config.RunnerParamsYCSB, l *slog.Logger) (runner.Runner, error) {
	return &ycsb{
		p: params,
		l: l,
	}, nil
}

// parseOutput parses go-ycsb JSON output.
func parseOutput(r io.Reader) (map[string]map[string]float64, error) {
	var res map[string]map[string]float64

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "write exception:") {
			return nil, errors.New("unexpected output")
		}

		if !strings.HasPrefix(line, "[{") {
			continue
		}

		var ms []measurement
		if err := json.Unmarshal([]byte(line), &ms); err != nil {
			return nil, err
		}

		res = make(map[string]map[string]float64)
		for _, m := range ms {
			res[strings.ToLower(m.Operation)] = map[string]float64{
				"takes":    m.TakesS,
				"count":    float64(m.Count),
				"ops":      m.OPS,
				"avg":      (time.Duration(m.AvgUs) * time.Microsecond).Seconds(),
				"min":      (time.Duration(m.MinUs) * time.Microsecond).Seconds(),
				"max":      (time.Duration(m.MaxUs) * time.Microsecond).Seconds(),
				"perc50":   (time.Duration(m.Perc50Us) * time.Microsecond).Seconds(),
				"perc90":   (time.Duration(m.Perc90Us) * time.Microsecond).Seconds(),
				"perc95":   (time.Duration(m.Perc95Us) * time.Microsecond).Seconds(),
				"perc99":   (time.Duration(m.Perc99Us) * time.Microsecond).Seconds(),
				"perc999":  (time.Duration(m.Perc999Us) * time.Microsecond).Seconds(),
				"perc9999": (time.Duration(m.Perc9999Us) * time.Microsecond).Seconds(),
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return res, err
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

	ms, err := parseOutput(io.TeeReader(pipe, os.Stdout))
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		return nil, err
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
func (y *ycsb) Run(ctx context.Context) (map[string]config.TestResult, error) {
	bin := filepath.Join("..", "bin", "go-ycsb")
	if _, err := os.Stat(bin); err != nil {
		return nil, err
	}

	bin, err := filepath.Abs(bin)
	if err != nil {
		return nil, err
	}

	args := []string{bin, "load", "mongodb", "-P", y.p.Args[0]}
	for _, p := range y.p.Args[1:] {
		args = append(args, "-p", p)
	}
	args = append(args, "-p", "outputstyle=json")

	y.l.InfoContext(ctx, "Load", slog.String("cmd", strings.Join(args, " ")))

	if _, err = run(ctx, args, y.p.Dir); err != nil {
		return nil, err
	}

	args[1] = "run"

	y.l.InfoContext(ctx, "Run", slog.String("cmd", strings.Join(args, " ")))

	return run(ctx, args, y.p.Dir)
}
