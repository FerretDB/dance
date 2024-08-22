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

// Package config provides project configuration.
package config

// Status represents the status of a single test.
type Status string

// Constants representing different expected or actual test statuses.
const (
	Fail    Status = "fail"
	Skip    Status = "skip"
	Pass    Status = "pass"
	Unknown Status = "unknown" // result can't be parsed
	Ignore  Status = "ignore"  // for fluky tests
)

// Config represents project configuration.
//
//nolint:vet // for readability
type Config struct {
	Runner  RunnerType
	Params  RunnerParams
	Results *ExpectedResults
}
