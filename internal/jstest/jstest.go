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
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/FerretDB/dance/internal"
)

const uri = "mongodb://host.docker.internal:27017"

func Run(ctx context.Context, args []string) (*internal.TestResults, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	ts := &internal.TestResults{}
	ts.TestResults = make(map[string]internal.TestResult)
	for _, testName := range args {
		output, err := runCommand("mongo", testName)
		if err != nil {
			if !strings.Contains(err.Error(), "exit status") {
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

	args = append([]string{"--verbose", "--norc", uri}, args...)
	dockerArgs := append([]string{"compose", "run", "-T", "--rm", command}, args...)
	cmd := exec.Command(bin, dockerArgs...)

	log.Printf("Running %q", strings.Join(dockerArgs, " "))
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %s", strings.Join(dockerArgs, " "), err)
	}

	return stdoutStderr, nil
}
