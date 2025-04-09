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

// Package command provides generic test runner.
package command

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/runner"
)

// command represents a generic test runner.
type command struct {
	p       *config.RunnerParamsCommand
	l       *slog.Logger
	verbose bool
}

// New creates a new `command` runner with given parameters.
func New(params *config.RunnerParamsCommand, l *slog.Logger, verbose bool) (runner.Runner, error) {
	return &command{
		p:       params,
		l:       l,
		verbose: verbose,
	}, nil
}

// execScripts stores the given shell script content in dir/file-XXX.sh and executes it.
// It returns the combined output of the script execution.
func execScript(ctx context.Context, dir, file, content string, verbose bool) ([]byte, error) {
	if dir == "" {
		dir = "."
	}

	f, err := os.CreateTemp(dir, file+"-*.sh")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	if verbose {
		content = "set -x\n\n" + content
	}

	content = "#!/bin/sh\n\n" + content + "\n"
	if _, err = f.WriteString(content); err != nil {
		return nil, err
	}

	if err = f.Chmod(0o755); err != nil {
		return nil, err
	}

	if err = f.Close(); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "./"+filepath.Base(f.Name()))
	cmd.Dir = dir

	var b runner.LockedBuffer

	if verbose {
		cmd.Stdout = io.MultiWriter(&b, os.Stdout)
		cmd.Stderr = io.MultiWriter(&b, os.Stderr)
	} else {
		cmd.Stdout = &b
		cmd.Stderr = &b
	}

	err = cmd.Run()
	return b.Bytes(), err
}

// Run implements [runner.Runner] interface.
func (c *command) Run(ctx context.Context) (res map[string]config.TestResult, err error) {
	var b []byte

	if c.p.Setup != "" {
		c.l.InfoContext(ctx, "Running setup")

		if b, err = execScript(ctx, c.p.Dir, "setup", c.p.Setup, c.verbose); err != nil {
			err = fmt.Errorf("%s\n%w", b, err)

			return
		}
	}

	if c.p.Teardown != "" {
		defer func() {
			c.l.InfoContext(ctx, "Running teardown")

			// canceled context should not prevent teardown
			if b, err = execScript(context.WithoutCancel(ctx), c.p.Dir, "teardown", c.p.Teardown, c.verbose); err != nil {
				err = fmt.Errorf("%s\n%w", b, err)
			}
		}()
	}

	res = c.runTests(ctx)

	return
}

// runTests executes tests and returns the results.
func (c *command) runTests(ctx context.Context) map[string]config.TestResult {
	res := make(map[string]config.TestResult, len(c.p.Tests))

	for _, t := range c.p.Tests {
		start := time.Now()
		c.l.InfoContext(ctx, "Running test", slog.String("test", t.Name))

		b, err := execScript(ctx, c.p.Dir, t.Name, t.Cmd, c.verbose)

		tc := config.TestResult{
			Status: config.Pass,
			Output: string(b),
		}

		args := []any{slog.String("test", t.Name), slog.Duration("duration", time.Since(start))}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			c.l.WarnContext(ctx, "Test failed", args...)

			tc.Status = config.Fail
			tc.Output += "\n" + err.Error()
		} else {
			c.l.InfoContext(ctx, "Test passed", args...)
		}

		res[t.Name] = tc
	}

	return res
}

// check interfaces
var (
	_ runner.Runner = (*command)(nil)
)
