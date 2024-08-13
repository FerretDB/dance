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
	Database []string `help:"${help_database}" enum:"${enum_database}" short:"d"`

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

	if len(cli.Database) == 0 {
		cli.Database = maps.Keys(configload.DBs)
		slices.Sort(cli.Database)
	}

	for _, db := range cli.Database {
		uri := configload.DBs[db]
		u, err := url.Parse(uri)
		if err != nil {
			log.Fatal(err)
		}

		port, err := strconv.Atoi(u.Port())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Waiting for port %d for %s / %s to be up...", port, db, uri)

		if err := waitForPort(ctx, port); err != nil {
			log.Fatal(err)
		}
	}

	if len(cli.Config) == 0 {
		var err error
		if cli.Config, err = filepath.Glob("*.yml"); err != nil {
			log.Fatal(err)
		}
	}

	for i, cf := range cli.Config {
		cli.Config[i] = filepath.Base(cf)
	}

	log.Printf("Run project configs: %v", cli.Config)

	for _, cf := range cli.Config {
		for _, db := range cli.Database {
			log.Println(db, cf)

			c, err := configload.Load(cf, db)
			if err != nil {
				log.Fatal(err)
			}

			rl := l.With("config", cf, "db", db)

			var runner runner.Runner
			switch c.Runner {
			case config.RunnerTypeCommand:
				runner, err = command.New(c.Params.(*config.RunnerParamsCommand), rl)
			case config.RunnerTypeGoTest:
				fallthrough
			case config.RunnerTypeJSTest:
				fallthrough
			case config.RunnerTypeYCSB:
				fallthrough
			default:
				log.Fatalf("unknown runner: %q", c.Runner)
			}

			if err != nil {
				log.Fatal(err)
			}

			res, err := runner.Run(ctx)
			if err != nil {
				log.Fatal(err)
			}

			cmp, err := c.Results.Compare(res)
			if err != nil {
				log.Fatal(err)
			}

			logResult("Unexpectedly failed", cmp.XFailed)
			logResult("Unexpectedly skipped", cmp.XSkipped)
			logResult("Unexpectedly passed", cmp.XPassed)

			if cli.Verbose {
				logResult("Expectedly failed", cmp.Failed)
				logResult("Expectedly skipped", cmp.Skipped)
				logResult("Expectedly passed", cmp.Passed)
			}

			logResult("Unknown", cmp.Unknown)

			log.Printf("Unexpectedly failed: %d.", len(cmp.XFailed))
			log.Printf("Unexpectedly skipped: %d.", len(cmp.XSkipped))
			log.Printf("Unexpectedly passed: %d.", len(cmp.XPassed))
			log.Printf("Expectedly failed: %d.", len(cmp.Failed))
			log.Printf("Expectedly skipped: %d.", len(cmp.Skipped))
			log.Printf("Expectedly passed: %d.", len(cmp.Passed))
			log.Printf("Unknown: %d.", len(cmp.Unknown))

			expectedStats, err := yaml.Marshal(c.Results.Stats)
			if err != nil {
				log.Fatal(err)
			}
			actualStats, err := yaml.Marshal(cmp.Stats)
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

			totalRun := cmp.Stats.Failed + cmp.Stats.Skipped + cmp.Stats.Passed
			msg := fmt.Sprintf(
				"%.2f%% (%d/%d) tests passed.",
				float64(cmp.Stats.Passed)/float64(totalRun)*100,
				cmp.Stats.Passed,
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
}
