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

import "time"

type action string

const (
	// actionRun means the test has started running.
	actionRun action = "run"
	// actionPause means the test has been paused.
	actionPause action = "pause"
	// actionCont means the test has continued running.
	actionCont action = "cont"
	// actionPass means the test passed.
	actionPass action = "pass"
	// actionBench means the benchmark printed log output but did not fail.
	actionBench action = "bench"
	// actionFail means the test or benchmark failed.
	actionFail action = "fail"
	// actionOutput means the test printed output.
	actionOutput action = "output"
	// actionSkip means the test was skipped or the package contained no tests.
	actionSkip action = "skip"
)

type testEvent struct {
	Time           time.Time `json:"Time"`
	Action         action    `json:"Action"`
	Package        string    `json:"Package"`
	Test           string    `json:"Test"`
	Output         string    `json:"Output"`
	ElapsedSeconds float64   `json:"Elapsed"`
}

func (te testEvent) Elapsed() time.Duration {
	return time.Duration(te.ElapsedSeconds * float64(time.Second))
}
