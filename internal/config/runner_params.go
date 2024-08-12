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

package config

// RunnerParams is common interface for runner parameters.
//
//sumtype:decl
type RunnerParams interface {
	runnerParams() // seal for sumtype
}

// RunnerParamsCommand represents `command` runner parameters.
type RunnerParamsCommand struct {
	Dir   string
	Setup string
	Tests []RunnerParamsCommandTest
}

type RunnerParamsCommandTest struct {
	Name string
	Cmd  string
}

func (rp *RunnerParamsCommand) runnerParams() {}

// check interfaces
var (
	_ RunnerParams = (*RunnerParamsCommand)(nil)
)
