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

package internal

import "strings"

type Result string

const (
	Unknown Result = "UNKNOWN"
	Pass    Result = "PASS"
	Fail    Result = "FAIL"
	Skip    Result = "SKIP"
)

type TestResult struct {
	Result Result
	Output string
}

func (tr *TestResult) IndentedOutput() string {
	return strings.Replace(tr.Output, "\n", "\n\t", -1)
}

type Results struct {
	TestResults map[string]TestResult
}
