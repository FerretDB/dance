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
	"log/slog"
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
	p *config.RunnerParamsGoTest
	l *slog.Logger
}

// New creates a new `gotest` runner with given parameters.
func New(params *config.RunnerParamsGoTest, l *slog.Logger) (runner.Runner, error) {
	return &goTest{
		p: params,
		l: l,
	}, nil
}

// Run implements [runner.Runner] interface.
func (c *goTest) Run(ctx context.Context) (map[string]config.TestResult, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args := append([]string{"test", "-v", "-json", "-count=1"}, c.p.Args...)

	cmd := exec.CommandContext(ctx, "go", args...)

	c.l.InfoContext(ctx, "Running", slog.String("cmd", strings.Join(cmd.Args, " ")))

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return nil, nil
}

// check interfaces
var (
	_ runner.Runner = (*goTest)(nil)
)
