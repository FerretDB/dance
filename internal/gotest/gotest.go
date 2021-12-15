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

// Package gotest contains `go test` runner.
package gotest

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/FerretDB/dance/internal"
)

func Run(dir string, args []string, verbose bool) (*internal.Results, error) {
	args = append([]string{"test", "-v", "-json", "-count=1"}, args...)
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
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

	res := &internal.Results{
		TestResults: make(map[string]internal.TestResult),
	}

	for {
		var event TestEvent
		if err = d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		testName := event.Package + "/" + event.Test
		result := res.TestResults[testName]
		result.Output += event.Output
		switch event.Action {
		case ActionPass:
			result.Result = internal.Pass
		case ActionFail:
			result.Result = internal.Fail
		case ActionSkip:
			result.Result = internal.Skip
		case ActionBench, ActionCont, ActionOutput, ActionPause, ActionRun:
			fallthrough
		default:
			result.Result = internal.Unknown
		}
		res.TestResults[testName] = result
	}

	if err = cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			err = nil
		}
	}

	return res, err
}
