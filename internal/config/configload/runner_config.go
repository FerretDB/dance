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

type runnerConfig interface {
	GetDir() string
	GetArgs() []string
	runnerConfig()
}

type runnerConfigCommand struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rc *runnerConfigCommand) GetDir() string    { return rc.Dir }
func (rc *runnerConfigCommand) GetArgs() []string { return rc.Args }
func (*runnerConfigCommand) runnerConfig()        {}

type runnerConfigGoTest struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rc *runnerConfigGoTest) GetDir() string    { return rc.Dir }
func (rc *runnerConfigGoTest) GetArgs() []string { return rc.Args }
func (*runnerConfigGoTest) runnerConfig()        {}

type runnerConfigJSTest struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rc *runnerConfigJSTest) GetDir() string    { return rc.Dir }
func (rc *runnerConfigJSTest) GetArgs() []string { return rc.Args }
func (*runnerConfigJSTest) runnerConfig()        {}

type runnerConfigYCSB struct {
	Dir  string   `yaml:"dir"`
	Args []string `yaml:"args"`
}

func (rc *runnerConfigYCSB) GetDir() string    { return rc.Dir }
func (rc *runnerConfigYCSB) GetArgs() []string { return rc.Args }
func (*runnerConfigYCSB) runnerConfig()        {}
