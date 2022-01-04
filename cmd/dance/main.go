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
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/exp/maps"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal"
	"github.com/FerretDB/dance/internal/gotest"
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
	log.SetFlags(0)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), unix.SIGTERM, unix.SIGINT)
	go func() {
		<-ctx.Done()
		log.Print("Stopping...")
		stop()
	}()

	log.Printf("Waiting for port 27017 to be up...")
	if err := waitForPort(ctx, 27017); err != nil {
		log.Fatal(err)
	}

	matches, err := filepath.Glob("*.yml")
	if err != nil {
		log.Fatal(err)
	}

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

		config, err := internal.LoadConfig(match)
		if err != nil {
			log.Fatal(err)
		}

		expectedConfig, err := config.Results.ForDB(*dbF)
		if err != nil {
			log.Fatal(err)
		}

		runRes, err := gotest.Run(ctx, dir, config.Args, *vF)
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
