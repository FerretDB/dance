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

package jstest

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const uri = "mongodb://host.docker.internal:27017"

func Run(ctx context.Context, args []string) error {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	// bad
	files := []string{}
	for _, f := range args {
		if strings.HasSuffix(f, ".js") {
			wd, _ := os.Getwd()
			ff := filepath.Join(wd, "mongo", f)
			matches, err := filepath.Glob(ff)
			if err != nil {
				return err
			}

			for _, m := range matches {
				i := strings.LastIndex(m, "/tests")
				m = m[i+1:]
				files = append(files, m)
			}
		}
	}

	return runCommand("mongo", uri, strings.Join(files, " "))
}

// runCommand runs command with args inside the mongo container.
func runCommand(command string, args ...string) error {
	bin, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	dockerArgs := append([]string{"compose", "run", "-T", "--rm", command}, args...)
	cmd := exec.Command(bin, dockerArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running %q", strings.Join(dockerArgs, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(dockerArgs, " "), err)
	}

	return nil
}
