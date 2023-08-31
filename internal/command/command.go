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

// Package command runs generic test commands.
package command

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/FerretDB/dance/internal/config"
)

// Run executes a generic test command with arguments in a specified directory.
// It captures the combined output and determines the test result based on the exit code.
func Run(ctx context.Context, dir string, args []string) (*config.TestResults, error) {
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
