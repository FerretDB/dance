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

package jstest

import (
	"context"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/FerretDB/dance/internal"
)

func Run(ctx context.Context, dir string, args []string) (*internal.TestResults, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	ts := &internal.TestResults{}
	ts.TestResults = make(map[string]internal.TestResult)

	files := []string{}

	for _, f := range args {
		matches, err := filepath.Glob(f)
		if err != nil {
			return nil, err
		}

		files = append(files, matches...)
	}

	for _, testName := range files {
		output, err := runCommand("mongo", testName)
		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				return nil, err
			}
		}

		if err == nil {
			ts.TestResults[testName] = internal.TestResult{
				Status: internal.Pass,
				Output: string(output),
			}

			continue
		}

		ts.TestResults[testName] = internal.TestResult{
			Status: internal.Fail,
			Output: string(output),
		}
	}

	return ts, nil
}

// runCommand runs command with args inside the mongo container and returns the
// combined output.
func runCommand(command string, args ...string) ([]byte, error) {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return nil, err
	}

	args = append([]string{"--verbose", "--norc", "mongodb://host.docker.internal:27017/"}, args...)
	dockerArgs := append([]string{"compose", "run", "-T", "--rm", command}, args...)
	cmd := exec.Command(bin, dockerArgs...)

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}
