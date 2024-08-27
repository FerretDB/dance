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

// Package gotest contains `gotest` runner.
package gotest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/runner"
)

// testEvent represents a single even emitted by `go test -json`.
//
// See https://pkg.go.dev/cmd/test2json#hdr-Output_Format.
type testEvent struct {
	Time           time.Time `json:"Time"`
	Action         string    `json:"Action"`
	Package        string    `json:"Package"`
	Test           string    `json:"Test"`
	Output         string    `json:"Output"`
	ElapsedSeconds float64   `json:"Elapsed"`
}

// Elapsed returns an elapsed time.
func (te testEvent) Elapsed() time.Duration {
	return time.Duration(te.ElapsedSeconds * float64(time.Second))
}

// goTest represents `gotest` runner.
type goTest struct {
	p       *config.RunnerParamsGoTest
	l       *slog.Logger
	verbose bool
}

// New creates a new `gotest` runner with given parameters.
func New(params *config.RunnerParamsGoTest, l *slog.Logger, verbose bool) (runner.Runner, error) {
	return &goTest{
		p:       params,
		l:       l,
		verbose: verbose,
	}, nil
}

// Run implements [runner.Runner] interface.
func (c *goTest) Run(ctx context.Context) (map[string]config.TestResult, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args := append([]string{"test", "-v", "-json", "-count=1"}, c.p.Args...)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = c.p.Dir
	cmd.Stderr = os.Stderr

	p, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	c.l.InfoContext(ctx, "Running", slog.String("cmd", strings.Join(cmd.Args, " ")))

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	d := json.NewDecoder(p)
	d.DisallowUnknownFields()

	res := make(map[string]config.TestResult)

	for {
		var event testEvent
		if err = d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if c.verbose {
			c.l.DebugContext(ctx, "", slog.Any("event", event))
		}

		if event.Test == "" {
			continue
		}

		testName := event.Package + "/" + event.Test

		result := res[testName]
		if result.Status == "" {
			result.Status = config.Unknown
		}

		result.Output += event.Output

		switch event.Action {
		case "fail":
			result.Status = config.Fail
		case "skip":
			result.Status = config.Skip
		case "pass":
			result.Status = config.Pass
		case "start", "run", "pause", "cont", "output", "bench":
			fallthrough
		default:
			result.Status = config.Unknown
		}

		res[testName] = result
	}

	err = cmd.Wait()
	c.l.InfoContext(ctx, "Done", slog.String("cmd", strings.Join(cmd.Args, " ")), slog.Any("err", err))

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.Exited() {
		err = nil
	}

	return res, err
}

// check interfaces
var (
	_ runner.Runner = (*goTest)(nil)
)
