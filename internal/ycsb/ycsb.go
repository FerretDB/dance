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

// Package ycsb contains `ycsb` runner.
package ycsb

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/FerretDB/dance/internal/config"
)

// Run runs `go-ycsb`.
//
// It loads and runs a YCSB workload.
// Properties defined in the YAML file will override properties defined in the workload parameter file.
func Run(ctx context.Context, dir string, args []string) (*config.TestResults, error) {
	bin := filepath.Join("..", "bin", "go-ycsb")
	if _, err := os.Stat(bin); err != nil {
		return nil, err
	}

	// because we set cmd.Dir, the relative path here is different
	bin = filepath.Join("..", bin)

	// load workload

	cliArgs := []string{"load", "mongodb", "-P", args[0]}
	for _, p := range args[1:] {
		cliArgs = append(cliArgs, "-p", p)
	}

	mongodbURI := "mongodb.url=" + os.Getenv("MONGODB_URI")

	cliArgs = append(cliArgs, "-p", mongodbURI)

	cmd := exec.CommandContext(ctx, bin, cliArgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// run workload with almost the same args

	cliArgs[0] = "run"

	cmd = exec.CommandContext(ctx, bin, cliArgs...)
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
