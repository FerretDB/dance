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
	"strings"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/dance/internal"
)

// Run runs `mongo`.
// Args is a list of filepath.Glob file patterns with additional support for !exclude.
func Run(ctx context.Context, dir string, args []string) (*internal.TestResults, error) {
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

	log.Printf("Total number of tests to run %d\n", len(files))

	res := &internal.TestResults{
		TestResults: make(map[string]internal.TestResult),
	}

	type item struct {
		file string
		err  error
		out  []byte
	}

	// tokens is a counting semaphore used to enforce a limit of 20 concurrent calls to runShellWithScript.
	tokens := make(chan struct{}, 20)

	ch := make(chan *item, len(files))

	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)

		go func(f string) {
			defer wg.Done()

			tokens <- struct{}{}
			it := &item{
				file: f,
			}

			// we set working_dir so use a relative path here instead
			rel, err := filepath.Rel("mongo", f)
			if err != nil {
				panic(err)
			}

			it.out, it.err = runShellWithScript(dir, rel)
			ch <- it

			<-tokens // release the token
		}(f)
	}

	wg.Wait()

	close(ch)

	for it := range ch {
		if it.err != nil {
			var exitErr *exec.ExitError
			if !errors.As(it.err, &exitErr) {
				return nil, it.err
			}

			res.TestResults[it.file] = internal.TestResult{
				Status: internal.Fail,
				Output: string(it.out),
			}
			continue
		}

		res.TestResults[it.file] = internal.TestResult{
			Status: internal.Pass,
			Output: string(it.out),
		}
	}

	return res, nil
}

// runShellWithScript runs the mongo shell inside a container with script and returns the combined output.
func runShellWithScript(dir, script string) ([]byte, error) {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return nil, err
	}

	dockerArgs := []string{"compose", "run", "-T", "--rm", "mongo"}
	shellArgs := []string{"--verbose", "--norc", "mongodb://host.docker.internal:27017/", "--eval", evalBuilder(script, nil), script}
	dockerArgs = append(dockerArgs, shellArgs...)

	cmd := exec.Command(bin, dockerArgs...)
	cmd.Dir = dir

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	return cmd.CombinedOutput()
}

// evalBuilder creates the TestData object and sets the testName property for the shell.
// It optionally sets additional properties on the TestData object.
func evalBuilder(script string, obj map[string]string) string {
	var eb bytes.Buffer
	eb.WriteString("TestData = new Object();")
	eb.WriteByte(' ')
	scriptName := filepath.Base(script)
	fmt.Fprintf(&eb, "TestData.testName = %q;", strings.TrimSuffix(scriptName, filepath.Ext(scriptName)))
	if obj == nil {
		return eb.String()
	}

	eb.WriteByte(' ')

	for key, value := range obj {
		fmt.Fprintf(&eb, "TestData.%s = %s;", key, value)
		eb.WriteByte(' ')
	}

	return eb.String()
}
