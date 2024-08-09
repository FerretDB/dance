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

package configload

// runnerParams is common interface for runner parameters.
//
//sumtype:decl
type runnerParams interface {
	GetDir() string
	GetArgs() []string

	runnerParams() // seal for sumtype
}

// runnerParamsCommand represents `command` runner parameters in the YAML project configuration file.
type runnerParamsCommand struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rp *runnerParamsCommand) GetDir() string    { return rp.Dir }
func (rp *runnerParamsCommand) GetArgs() []string { return rp.Args }
func (rp *runnerParamsCommand) runnerParams()     {}

// runnerParamsGoTest represents `gotest` runner parameters in the YAML project configuration file.
type runnerParamsGoTest struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rp *runnerParamsGoTest) GetDir() string    { return rp.Dir }
func (rp *runnerParamsGoTest) GetArgs() []string { return rp.Args }
func (rp *runnerParamsGoTest) runnerParams()     {}

// runnerParamsJSTest represents `jstest` runner parameters in the YAML project configuration file.
type runnerParamsJSTest struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rp *runnerParamsJSTest) GetDir() string    { return rp.Dir }
func (rp *runnerParamsJSTest) GetArgs() []string { return rp.Args }
func (rp *runnerParamsJSTest) runnerParams()     {}

// runnerParamsYCSB represents `ycsb` runner parameters in the YAML project configuration file.
type runnerParamsYCSB struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rp *runnerParamsYCSB) GetDir() string    { return rp.Dir }
func (rp *runnerParamsYCSB) GetArgs() []string { return rp.Args }
func (rp *runnerParamsYCSB) runnerParams()     {}

// check interfaces
var (
	_ runnerParams = (*runnerParamsCommand)(nil)
	_ runnerParams = (*runnerParamsGoTest)(nil)
	_ runnerParams = (*runnerParamsJSTest)(nil)
	_ runnerParams = (*runnerParamsYCSB)(nil)
)
