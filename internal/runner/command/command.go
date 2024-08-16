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

// Package command contains generic test runner.
package command

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/runner"
)

// command represents a generic test runner.
type command struct {
	p *config.RunnerParamsCommand
	l *slog.Logger
}

// New creates a new `command` runner with given parameters.
func New(params *config.RunnerParamsCommand, l *slog.Logger) (runner.Runner, error) {
	return &command{
		p: params,
		l: l,
	}, nil
}

// execScripts stores the given shell script content in dir/file-XXX.sh and executes it.
// It returns the combined output of the script execution.
func execScript(ctx context.Context, dir, file, content string) ([]byte, error) {
	f, err := os.CreateTemp(dir, file+"-*.sh")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

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

	return cmd.CombinedOutput()
}

// Run implements [runner.Runner] interface.
func (c *command) Run(ctx context.Context) (map[string]config.TestResult, error) {
	if c.p.Setup != "" {
		c.l.InfoContext(ctx, "Running setup")

		b, err := execScript(ctx, c.p.Dir, c.p.Dir, c.p.Setup)
		if err != nil {
			return nil, fmt.Errorf("%s\n%w", b, err)
		}
	}

	res := make(map[string]config.TestResult, len(c.p.Tests))

	for _, t := range c.p.Tests {

		c.l.InfoContext(ctx, "Running test", slog.String("test", t.Name))

		b, err := execScript(ctx, c.p.Dir, t.Name, t.Cmd)

		tc := config.TestResult{
			Status: config.Pass,
			Output: string(b),
		}

		if err != nil {
			c.l.WarnContext(ctx, "Test failed", slog.String("test", t.Name), slog.String("error", err.Error()))

			tc.Status = config.Fail
			tc.Output += "\n" + err.Error()
		} else {
			c.l.InfoContext(ctx, "Test passed", slog.String("test", t.Name))
		}

		res[t.Name] = tc
	}

	return res, nil
}

// check interfaces
var (
	_ runner.Runner = (*command)(nil)
)
