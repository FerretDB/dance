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
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// parseOutput parses mongo-bench output.
func parseOutput(r io.Reader) (map[string]map[string]float64, error) {
	var res map[string]map[string]float64

	return res, errors.New("unimplemented")
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
