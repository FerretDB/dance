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

	"github.com/FerretDB/dance/internal"
)

// Run runs generic test.
// It runs a command with arguments in a directory and returns the combined output as is.
// If the command exits with a non-zero exit code, the test fails.
func Run(ctx context.Context, dir string, args []string) (*internal.TestResults, error) {
	bin, err := exec.LookPath(args[0])
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, bin, args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	res := &internal.TestResults{
		TestResults: map[string]internal.TestResult{
			dir: {
				Status: internal.Pass,
			},
		},
	}

	if err = cmd.Run(); err != nil {
		res.TestResults[dir] = internal.TestResult{
			Status: internal.Fail,
			Output: err.Error(),
		}
	}

	return res, nil
}
