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
	"log/slog"
	"os"
	"os/exec"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/runner"
)

type command struct {
	p *config.RunnerParamsCommand
	l *slog.Logger
}

// Run runs generic test.
// It runs a command with arguments in a directory and returns the combined output as is.
// If the command exits with a non-zero exit code, the test fails.
func New(params *config.RunnerParamsCommand, l *slog.Logger) (runner.Runner, error) {
	return &command{
		p: params,
		l: l,
	}, nil
}

func execScript(ctx context.Context, dir, file, content string) error {
	f, err := os.CreateTemp(dir, file+"-*.sh")
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	content = "#!/bin/sh\n\n" + content + "\n"
	if _, err = f.WriteString(content); err != nil {
		return err
	}

	if err = f.Chmod(0o755); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "./"+f.Name())
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Run implements [runner.Runner] interface.
func (c *command) Run(ctx context.Context) (map[string]config.TestResult, error) {
	c.l.InfoContext(ctx, "Running setup")

	if err := execScript(ctx, c.p.Dir, c.p.Dir, c.p.Setup); err != nil {
		return nil, err
	}

	res := make(map[string]config.TestResult, len(c.p.Tests))

	for _, t := range c.p.Tests {
		tc := config.TestResult{
			Status: config.Pass,
		}

		c.l.InfoContext(ctx, "Running test", slog.String("test", t.Name))

		if err := execScript(ctx, c.p.Dir, t.Name, t.Cmd); err != nil {
			tc.Status = config.Fail
			tc.Output = err.Error()
		}

		res[t.Name] = tc
	}

	return res, nil
}
