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
	"os/exec"
	"strings"

	"github.com/FerretDB/dance/internal"
)

// Run runs YCSB workloads.
func Run(ctx context.Context, dir string, args []string) (*internal.TestResults, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	res := &internal.TestResults{
		TestResults: make(map[string]internal.TestResult),
	}

	out, err := runWorkload(dir, args...)
	if err != nil {
		return nil, err
	}

	// just log the output of the workload for now
	log.Println(string(out))

	res.TestResults[dir] = internal.TestResult{
		Status: internal.Pass,
		Output: string(out),
	}

	return res, nil
}

// runWorkload runs a YCSB workload.
func runWorkload(dir string, args ...string) ([]byte, error) {
	bin, err := exec.LookPath("go-ycsb")
	if err != nil {
		return nil, err
	}

	wlArgs := append(
		[]string{
			"load",
			"mongodb",
			"-P",
		}, args...,
	)

	// TODO: create a directory tests/ycsb/workloads and put the workloads there
	wlArgs = append(wlArgs, "-p", "mongodb.url=mongodb://localhost:27017/")
	cmd := exec.Command(bin, wlArgs...)
	cmd.Dir = dir

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}
