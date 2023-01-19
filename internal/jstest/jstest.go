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
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/dance/internal"
)

// Run runs jstests.
func Run(ctx context.Context, dir string, args []string) (*internal.TestResults, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	filesM := make(map[string]struct{})

	// remove duplicates if globs match same files
	for _, f := range args {
		matches, err := filepath.Glob(f)
		if err != nil {
			return nil, err
		}

		for _, m := range matches {
			filesM[m] = struct{}{}
		}
	}

	files := maps.Keys(filesM)
	sort.Strings(files)

	res := &internal.TestResults{
		TestResults: make(map[string]internal.TestResult),
	}

	volume := "tests"
	for _, testName := range files {
		output, err := runCommand(dir, "mongo", filepath.Join(volume, testName))
		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				return nil, err
			}
		}

		if err == nil {
			res.TestResults[testName] = internal.TestResult{
				Status: internal.Pass,
				Output: string(output),
			}

			continue
		}

		res.TestResults[testName] = internal.TestResult{
			Status: internal.Fail,
			Output: string(output),
		}
	}

	return res, nil
}

// runCommand runs command with args inside the mongo container and returns the
// combined output.
func runCommand(dir, command string, args ...string) ([]byte, error) {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return nil, err
	}

	args = append([]string{"--verbose", "--norc", "mongodb://host.docker.internal:27017/"}, args...)
	dockerArgs := append([]string{"compose", "run", "-T", "--rm", command}, args...)
	cmd := exec.Command(bin, dockerArgs...)
	cmd.Dir = dir

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}
