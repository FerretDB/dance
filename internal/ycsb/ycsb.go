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

// Package ycsb contains ycsb runner.
package ycsb

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/FerretDB/dance/internal"
)

// Run runs YCSB workloads.
// It loads and runs a YCSB workload. Properties defined in the YAML file
// will override properties defined in the workload parameter file.
func Run(ctx context.Context, dir string, args []string) (*internal.TestResults, error) {
	res := &internal.TestResults{
		TestResults: make(map[string]internal.TestResult),
	}

	_, err := os.Stat("../bin/go-ycsb")
	if err != nil {
		return nil, err
	}

	// because we set cmd.Dir, the relative path here is different
	bin := "../../bin/go-ycsb"

	wlFile := args[0]
	wlArgs := []string{"load", "mongodb", "-P", wlFile}
	wlArgs = append(wlArgs, "-p")
	wlArgs = append(wlArgs, args[1:]...)

	cmd := exec.CommandContext(ctx, bin, wlArgs...)
	cmd.Dir = dir

	log.Printf("Loading workload with properties %s", strings.Join(args, " "))

	// load the workload
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	// the run phase will execute the workload and will report performance statistics on stdout
	cmd.Args[1] = "run"
	cmd = exec.CommandContext(ctx, bin, cmd.Args[1:]...)
	cmd.Dir = dir

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	log.Print(string(out))

	res.TestResults[dir] = internal.TestResult{
		Status: internal.Pass,
		Output: "",
	}

	return res, nil
}
