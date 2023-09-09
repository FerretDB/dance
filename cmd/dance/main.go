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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal/command"
	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/configload"
	"github.com/FerretDB/dance/internal/gotest"
	"github.com/FerretDB/dance/internal/jstest"
	"github.com/FerretDB/dance/internal/ycsb"
)

func waitForPort(ctx context.Context, port uint16) error {
	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			conn.Close()
			return nil
		}

		sleepCtx, sleepCancel := context.WithTimeout(ctx, time.Second)
		<-sleepCtx.Done()
		sleepCancel()
	}

	return ctx.Err()
}

func logResult(label string, res map[string]string) {
	keys := maps.Keys(res)
	if len(keys) == 0 {
		return
	}

	log.Printf("%s tests:", label)
	sort.Strings(keys)
	for _, t := range keys {
		out := res[t]
		log.Printf("%s:\n\t%s", t, out)
	}
}

func main() {
	dbF := flag.String("db", "", "database to use: ferretdb, mongodb")
	vF := flag.Bool("v", false, "be verbose")
	pF := flag.Int("p", 0, "number of tests to run in parallel")
	log.SetFlags(0)
	flag.Parse()

	// TODO https://github.com/FerretDB/dance/issues/30
	if *dbF == "" {
		log.Fatal("-db is required")
	}

	ctx, stop := notifyAppTermination(context.Background())
	go func() {
		<-ctx.Done()
		log.Print("Stopping...")
		stop()
	}()

	const port = 27017
	log.Printf("Waiting for port %d to be up...", port)

	if err := waitForPort(ctx, port); err != nil {
		log.Fatal(err)
	}

	matches, err := filepath.Glob("*.yml")
	if err != nil {
		log.Fatal(err)
	}

	// TODO validate that args are in matches
	if flag.NArg() != 0 {
		matches = matches[:0:cap(matches)]
		for _, arg := range flag.Args() {
			matches = append(matches, arg+".yml")
		}
	}

	log.Printf("Run configurations: %v", matches)

	for _, match := range matches {
		dir := strings.TrimSuffix(match, filepath.Ext(match))
		log.Printf("%s (%s)", match, dir)

		cfg, err := configload.Load(match)
		if err != nil {
			log.Fatal(err)
		}

		if cfg.Dir != "" {
			dir = cfg.Dir
			log.Printf("\tDir changed to %s", dir)
		}

		expectedConfig, err := cfg.ForDB(*dbF)
		if err != nil {
			log.Fatal(err)
		}

		var runRes *config.TestResults

		switch cfg.Runner {
		case config.RunnerTypeCommand:
			runRes, err = command.Run(ctx, dir, cfg.Args)
		case config.RunnerTypeGoTest:
			runRes, err = gotest.Run(ctx, dir, cfg.Args, *vF, *pF)
		case config.RunnerTypeJSTest:
			runRes, err = jstest.Run(ctx, dir, cfg.Args, *pF)
		case config.RunnerTypeYCSB:
			runRes, err = ycsb.Run(ctx, dir, cfg.Args)
		default:
			log.Fatalf("unknown runner: %q", cfg.Runner)
		}

		if err != nil {
			log.Fatal(err)
		}

		compareRes, err := expectedConfig.Compare(runRes)
		if err != nil {
			log.Fatal(err)
		}

		keys := maps.Keys(compareRes.UnexpectedRest)
		if len(keys) != 0 {
			log.Printf("Unexpected/unknown results:")
			sort.Strings(keys)
			for _, t := range keys {
				res := compareRes.UnexpectedRest[t]
				log.Printf("%s %s:\n\t%s", t, res.Status, res.IndentedOutput())
			}
		}

		logResult("Unexpectedly failed", compareRes.UnexpectedFail)
		logResult("Unexpectedly skipped", compareRes.UnexpectedSkip)
		logResult("Unexpectedly passed", compareRes.UnexpectedPass)

		if *vF {
			logResult("Expectedly failed", compareRes.ExpectedFail)
			logResult("Expectedly skipped", compareRes.ExpectedSkip)
			logResult("Expectedly passed", compareRes.ExpectedPass)
		}

		log.Printf("Unexpected/unknown results: %d.", len(compareRes.UnexpectedRest))
		log.Printf("Unexpectedly failed: %d.", len(compareRes.UnexpectedFail))
		log.Printf("Unexpectedly skipped: %d.", len(compareRes.UnexpectedSkip))
		log.Printf("Unexpectedly passed: %d.", len(compareRes.UnexpectedPass))
		log.Printf("Expectedly failed: %d.", len(compareRes.ExpectedFail))
		log.Printf("Expectedly skipped: %d.", len(compareRes.ExpectedSkip))
		log.Printf("Expectedly passed: %d.", len(compareRes.ExpectedPass))

		expectedStats, err := yaml.Marshal(expectedConfig.Stats)
		if err != nil {
			log.Fatal(err)
		}
		actualStats, err := yaml.Marshal(compareRes.Stats)
		if err != nil {
			log.Fatal(err)
		}

		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(expectedStats)),
			B:        difflib.SplitLines(string(actualStats)),
			FromFile: "Expected",
			ToFile:   "Actual",
			Context:  10,
		})
		if err != nil {
			log.Fatal(err)
		}
		if diff != "" {
			log.Fatalf("\nUnexpected stats:\n%s", diff)
		}
	}
}
