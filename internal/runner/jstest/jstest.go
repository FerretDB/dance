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

// Package jstest contains `mongo` runner.
package jstest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/dance/internal/config"
)

// Run runs `mongo`.
// Args is a list of filepath.Glob file patterns with additional support for !exclude.
func Run(ctx context.Context, dir string, args []string, parallel int) (map[string]config.TestResult, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	filesM := make(map[string]struct{})

	excludeFilesM := make(map[string]struct{})

	// remove duplicates if globs match same files
	for _, f := range args {
		if exclude, found := strings.CutPrefix(f, "!"); found {
			matches, err := filepath.Glob(exclude)
			if err != nil {
				return nil, err
			}

			// exclude files from the list of files to run
			for _, m := range matches {
				excludeFilesM[m] = struct{}{}
			}

			continue
		}

		matches, err := filepath.Glob(f)
		if err != nil {
			return nil, err
		}

		for _, m := range matches {
			// skip excluded files
			if _, ok := excludeFilesM[m]; ok {
				continue
			}
			filesM[m] = struct{}{}
		}
	}

	files := maps.Keys(filesM)

	res := make(map[string]config.TestResult, len(files))

	type item struct {
		file string
		err  error
		out  []byte
	}

	input := make(chan string, len(files))
	output := make(chan *item, len(files))

	if parallel <= 0 {
		parallel = runtime.NumCPU()
	}

	log.Printf("Running up to %d tests in parallel.", parallel)

	for i := 0; i < parallel; i++ {
		go func() {
			for f := range input {
				it := &item{
					file: f,
				}

				// we set working_dir so use a relative path here instead
				rel, err := filepath.Rel("mongo", f)
				if err != nil {
					panic(err)
				}

				it.out, it.err = runMongo(dir, rel)
				output <- it
			}
		}()
	}

	for _, f := range files {
		input <- f
	}

	close(input)

	for i := 0; i < len(files); i++ {
		it := <-output

		if it.err != nil {
			var exitErr *exec.ExitError
			if !errors.As(it.err, &exitErr) {
				return nil, it.err
			}

			res[it.file] = config.TestResult{
				Status: config.Fail,
				Output: string(it.out),
			}
			continue
		}

		res[it.file] = config.TestResult{
			Status: config.Pass,
			Output: string(it.out),
		}
	}

	return res, nil
}

// runMongo runs the mongo shell inside a container with file and returns the combined output.
func runMongo(dir, file string) ([]byte, error) {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return nil, err
	}

	dockerArgs := []string{"compose", "run", "-T", "--rm", "mongo"}
	shellArgs := []string{
		"--verbose", "--norc", "mongodb://host.docker.internal:27017/",
		"--eval", evalBuilder(file), file,
	}
	dockerArgs = append(dockerArgs, shellArgs...)

	cmd := exec.Command(bin, dockerArgs...)
	cmd.Dir = dir

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}

// evalBuilder creates the TestData object and sets the testName property for the shell.
func evalBuilder(file string) string {
	var eb bytes.Buffer
	fmt.Fprintf(&eb, "TestData = new Object(); ")
	fileName := filepath.Base(file)
	fmt.Fprintf(&eb, "TestData.testName = %q;", strings.TrimSuffix(fileName, filepath.Ext(fileName)))

	return eb.String()
}
