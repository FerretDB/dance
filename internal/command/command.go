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
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/FerretDB/dance/internal/config"
)

// Run runs generic test.
// It runs a command with arguments in a directory and returns the combined output as is.
// If the command exits with a non-zero exit code, the test fails.
func Run(ctx context.Context, dir string, args []string) (*config.TestResults, error) {
	allCommands := true

	for _, arg := range args {
		if !strings.HasSuffix(arg, ".sh") {
			allCommands = false
		}
	}

	if allCommands {
		return runAllAsCommands(ctx, dir, args)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	res := &config.TestResults{
		TestResults: map[string]config.TestResult{
			dir: {
				Status: config.Pass,
			},
		},
	}

	if err := cmd.Run(); err != nil {
		res.TestResults[dir] = config.TestResult{
			Status: config.Fail,
			Output: err.Error(),
		}
	}

	return res, nil
}

func runAllAsCommands(ctx context.Context, dir string, files []string) (*config.TestResults, error) {
	res := &config.TestResults{
		TestResults: make(map[string]config.TestResult, len(files)),
	}

	for _, f := range files {
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", f)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("Running %s", strings.Join(cmd.Args, " "))

		res.TestResults[f] = config.TestResult{
			Status: config.Pass,
		}

		if err := cmd.Run(); err != nil {
			res.TestResults[f] = config.TestResult{
				Status: config.Fail,
				Output: err.Error(),
			}
		}
	}

	return res, nil
}
