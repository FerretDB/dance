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

type Test struct {
	Name string
	Cmd  string
}

type Params struct {
	Name     string
	Dir      string
	SetupCmd string
	Tests    []Test
	L        *slog.Logger
}

type command struct {
	Params
}

// Run runs generic test.
// It runs a command with arguments in a directory and returns the combined output as is.
// If the command exits with a non-zero exit code, the test fails.
func New(params Params) (runner.Runner, error) {
	return &command{
		Params: params,
	}, nil
}

func (c *command) execScript(ctx context.Context, name, content string) error {
	f, err := os.CreateTemp(c.Dir, name+"-*.sh")
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
	cmd.Dir = c.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (c *command) Setup(ctx context.Context) error {
	c.L.InfoContext(ctx, "Running setup", slog.String("project", c.Name))
	return c.execScript(ctx, c.Name, c.SetupCmd)
}

func (c *command) Run(ctx context.Context) (*config.TestResults, error) {
	res := &config.TestResults{
		TestResults: map[string]config.TestResult{},
	}

	for _, t := range c.Tests {
		tc := config.TestResult{
			Status: config.Pass,
		}

		c.L.InfoContext(ctx, "Running test", slog.String("project", c.Name), slog.String("test", t.Name))
		if err := c.execScript(ctx, t.Name, t.Cmd); err != nil {
			tc.Status = config.Fail
			tc.Output = err.Error()
		}

		res.TestResults[t.Name] = tc
	}

	return res, nil
}
