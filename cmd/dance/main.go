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
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/FerretDB/dance/internal/config"
	"github.com/FerretDB/dance/internal/configload"
	"github.com/FerretDB/dance/internal/runner"
	"github.com/FerretDB/dance/internal/runner/command"
)

func waitForPort(ctx context.Context, port int) error {
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

var cli struct {
	// TODO https://github.com/FerretDB/dance/issues/30
	Database string `help:"${help_database}" enum:"${enum_database}" required:"" short:"d"`

	Verbose bool `help:"Be more verbose." short:"v"`

	Config []string `arg:"" help:"Project configurations to run." type:"existingfile" optional:""`
}

func parseCLI() {
	dbs := maps.Keys(configload.DBs)
	slices.Sort(dbs)

	dbsHelp := make([]string, len(dbs))
	for i, db := range dbs {
		dbsHelp[i] = fmt.Sprintf("%s (%s)", db, configload.DBs[db])
	}

	kongOptions := []kong.Option{
		kong.Vars{
			"help_database": fmt.Sprintf("Database: %s.", strings.Join(dbsHelp, ", ")),
			"enum_database": strings.Join(dbs, ","),
		},
		kong.DefaultEnvars("DANCE"),
	}

	kong.Parse(&cli, kongOptions...)
}

func main() {
	log.SetFlags(0)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	l := slog.Default()

	parseCLI()

	ctx, stop := sigTerm(context.Background())

	go func() {
		<-ctx.Done()
		l.Info("Stopping...")

		// second SIGTERM should immediately stop the process
		stop()
	}()

	u, err := url.Parse(configload.DBs[cli.Database])
	if err != nil {
		log.Fatal(err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for port %d to be up...", port)

	if err := waitForPort(ctx, port); err != nil {
		log.Fatal(err)
	}

	configs := cli.Config
	if len(configs) == 0 {
		configs, err = filepath.Glob("*.yml")
		if err != nil {
			log.Fatal(err)
		}
	}

	for i, c := range configs {
		configs[i] = filepath.Base(c)
	}

	log.Printf("Run project configs: %v", configs)

	for _, c := range configs {
		log.Print(c)

		pc, err := configload.Load(c, cli.Database)
		if err != nil {
			log.Fatal(err)
		}

		var runRes map[string]config.TestResult
		var runner runner.Runner

		switch pc.Runner {
		case config.RunnerTypeCommand:
			runner, err = command.New(pc.Params.(*config.RunnerParamsCommand), l)
		case config.RunnerTypeGoTest:
			fallthrough
		case config.RunnerTypeJSTest:
			fallthrough
		case config.RunnerTypeYCSB:
			fallthrough
		default:
			log.Fatalf("unknown runner: %q", pc.Runner)
		}

		if err != nil {
			log.Fatal(err)
		}

		runRes, err = runner.Run(ctx)
		if err != nil {
			log.Fatal(err)
		}

		compareRes, err := pc.Results.Compare(runRes)
		if err != nil {
			log.Fatal(err)
		}

		logResult("Unexpectedly failed", compareRes.XFailed)
		logResult("Unexpectedly skipped", compareRes.XSkipped)
		logResult("Unexpectedly passed", compareRes.XPassed)

		if cli.Verbose {
			logResult("Expectedly failed", compareRes.Failed)
			logResult("Expectedly skipped", compareRes.Skipped)
			logResult("Expectedly passed", compareRes.Passed)
		}

		logResult("Unknown", compareRes.Unknown)

		log.Printf("Unexpectedly failed: %d.", len(compareRes.XFailed))
		log.Printf("Unexpectedly skipped: %d.", len(compareRes.XSkipped))
		log.Printf("Unexpectedly passed: %d.", len(compareRes.XPassed))
		log.Printf("Expectedly failed: %d.", len(compareRes.Failed))
		log.Printf("Expectedly skipped: %d.", len(compareRes.Skipped))
		log.Printf("Expectedly passed: %d.", len(compareRes.Passed))
		log.Printf("Unknown: %d.", len(compareRes.Unknown))

		expectedStats, err := yaml.Marshal(pc.Results.Stats)
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

		totalRun := compareRes.Stats.Failed + compareRes.Stats.Skipped + compareRes.Stats.Passed
		msg := fmt.Sprintf(
			"%.2f%% (%d/%d) tests passed.",
			float64(compareRes.Stats.Passed)/float64(totalRun)*100,
			compareRes.Stats.Passed,
			totalRun,
		)
		log.Print(msg)

		// Make percentage more visible on GitHub Actions.
		// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
		if os.Getenv("GITHUB_ACTIONS") == "true" {
			action := githubactions.New()
			action.Noticef("%s", msg)
		}
	}
}
