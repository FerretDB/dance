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

package gotest

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/FerretDB/dance/internal/config"
)

// Run runs `go test`.
// Args contain additional arguments to `go test`.
// `-v -json -p=1 -count=1` are always added.
// `-race` is added if possible.
func Run(ctx context.Context, dir string, args []string, verbose bool, parallel int) (map[string]config.TestResult, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	args = append([]string{"test", "-v", "-json", "-p=1", "-count=1"}, args...)

	// implicitly defaults to GOMAXPROCS
	if parallel > 0 {
		log.Printf("Running up to %d tests in parallel.", parallel)
		args = append(args, "-parallel="+strconv.Itoa(parallel))
	}

	// use the same condition as in FerretDB's Taskfile.yml
	if runtime.GOOS != "windows" && runtime.GOARCH != "arm" && runtime.GOARCH != "riscv64" {
		args = append(args, "-race")
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	p, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var r io.Reader = p
	if verbose {
		r = io.TeeReader(p, os.Stdout)
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	d := json.NewDecoder(r)
	d.DisallowUnknownFields()

	res := make(map[string]config.TestResult)

	for {
		var event testEvent
		if err = d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// skip package failures
		if event.Test == "" {
			continue
		}

		testName := event.Package + "/" + event.Test

		result := res[testName]
		if result.Status == "" {
			result.Status = config.Unknown
		}

		result.Output += event.Output

		switch event.Action {
		case "pass":
			result.Status = config.Pass
		case "fail":
			result.Status = config.Fail
		case "skip":
			result.Status = config.Skip
		case "start", "run", "pause", "cont", "output", "bench":
			fallthrough
		default:
			result.Status = config.Unknown
		}

		res[testName] = result
	}

	if err = cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			err = nil
		}
	}

	return res, err
}
