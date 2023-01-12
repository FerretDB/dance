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

// status represents single test status.
type status string

const (
	Pass     status = "pass"
	Skip     status = "skip"
	Fail     status = "fail"
	Unstable status = "unstable"
	Unknown  status = "unknown"
)

var knownStatuses = map[status]struct{}{
	Pass: {},
	Skip: {},
	Fail: {},
}

// TestResult represents single test result (status and output).
type TestResult struct {
	Status status
	Output string
}

func (tr *TestResult) IndentedOutput() string {
	return strings.ReplaceAll(tr.Output, "\n", "\n\t")
}

// TestResults represents results of a multiple tests.
//
// They are returned by runners.
type TestResults struct {
	// Test results by full test name.
	TestResults map[string]TestResult
}
