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

// Package service contains service tests runner.
package service

import (
	"context"
	"path/filepath"

	"github.com/FerretDB/dance/internal"
	"github.com/FerretDB/dance/internal/jstest"
)

// Run runs service tests.
func Run(ctx context.Context, dir string, args, excludeArgs []string) (*internal.TestResults, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	_ = dir
	dir = "mongo"

	excludeMatches := []string{}

	// inclusion patterns are only allowed so we do this
	for _, f := range excludeArgs {
		matches, err := filepath.Glob(f)
		if err != nil {
			return nil, err
		}

		excludeMatches = append(excludeMatches, matches...)
	}

	return jstest.Run(ctx, dir, args, excludeMatches)
}
