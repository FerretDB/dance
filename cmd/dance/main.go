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

	"golang.org/x/exp/maps"

	"github.com/FerretDB/dance/internal"
	"github.com/FerretDB/dance/internal/gotest"
)

func waitForPort(ctx context.Context, port uint16) error {
	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			conn.Close()

			// FIXME https://github.com/FerretDB/FerretDB/issues/92
			time.Sleep(time.Second)

			return nil
		}

		sleepCtx, sleepCancel := context.WithTimeout(ctx, time.Second)
		<-sleepCtx.Done()
		sleepCancel()
	}

	return ctx.Err()
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	ctx := context.Background()

	log.Printf("Waiting for port 27017 to be up...")
	if err := waitForPort(ctx, 27017); err != nil {
		log.Fatal(err)
	}

	matches, err := filepath.Glob("*.yml")
	if err != nil {
		log.Fatal(err)
	}

	for _, match := range matches {
		dir := strings.TrimSuffix(match, filepath.Ext(match))
		log.Printf("%s (%s)", match, dir)

		config, err := internal.LoadConfig(match)
		if err != nil {
			log.Fatal(err)
		}

		runRes, err := gotest.Run(dir, config.Args)
		if err != nil {
			log.Fatal(err)
		}

		compareRes := config.Tests.Compare(runRes)

		log.Printf("Unexpectedly failed tests (regressions):")
		keys := maps.Keys(compareRes.UnexpectedFail)
		sort.Strings(keys)
		for _, t := range keys {
			res := compareRes.UnexpectedFail[t]
			log.Printf("%s %s:\n\t%s", t, res.Result, res.IndentedOutput())
		}

		log.Printf("\nPassed tests:")
		for _, t := range compareRes.ExpectedPass {
			log.Print(t)
		}

		log.Printf("\nFailed tests:")
		keys = maps.Keys(compareRes.Fail)
		sort.Strings(keys)
		for _, t := range keys {
			res := compareRes.Fail[t]
			log.Printf("%s %s:\n\t%s", t, res.Result, res.IndentedOutput())
		}

		log.Printf("\nThe rest:")
		keys = maps.Keys(compareRes.Rest)
		sort.Strings(keys)
		for _, t := range keys {
			res := compareRes.Rest[t]
			log.Printf("%s %s:\n\t%s", t, res.Result, res.IndentedOutput())
		}

		if len(compareRes.UnexpectedFail) > 0 {
			log.Fatal("Unexpectedly failed tests present.")
		}
	}
}
