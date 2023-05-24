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

// Package command contains command tests runner.
package command

import (
	"context"
	"errors"
	"log"
	"os/exec"
	"strings"

	"github.com/FerretDB/dance/internal"
)

// Run runs generic command tests.
func Run(ctx context.Context, dir string, args []string) (*internal.TestResults, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	res := &internal.TestResults{
		TestResults: make(map[string]internal.TestResult),
	}

	_ = dir
	dir = args[len(args)-1]

	cmd := args[0]

	cmdArgs := args[1 : len(args)-1]

	for _, arg := range cmdArgs {
		out, err := runCommand(dir, cmd, arg)
		if err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				return nil, err
			}

			res.TestResults[dir] = internal.TestResult{
				Status: internal.Fail,
				Output: string(out),
			}

			continue
		}

		res.TestResults[dir] = internal.TestResult{
			Status: internal.Pass,
			Output: string(out),
		}
	}

	return res, nil
}

// runCommand runs command in dir with args returns the combined output.
func runCommand(dir, command string, args ...string) ([]byte, error) {
	bin, err := exec.LookPath(command)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}
