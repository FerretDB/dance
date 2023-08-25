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

// Config represents dance configuration.
//
//nolint:govet // we don't care about alignment there
type Config struct {
	Runner  string
	Dir     string
	Args    []string // meaning depends on runner type
	Results Results
}

// Results represents expected dance results.
type Results struct {
	// Expected results for both FerretDB and MongoDB.
	Common   *TestsConfig
	FerretDB *TestsConfig
	MongoDB  *TestsConfig
}
