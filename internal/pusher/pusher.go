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

// Package pusher provider pusher of test results to MongoDB-compatible database.
package pusher

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/dance/internal/config"
)

// Client represents a MongoDB client.
type Client struct {
	l            *slog.Logger
	c            *mongo.Client
	pingerCancel context.CancelFunc
	pingerDone   chan struct{}
	database     string
	hostname     string
	runner       string
	repository   string
}

// New creates a new MongoDB client with given URI.
func New(uri string, l *slog.Logger) (*Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		return nil, fmt.Errorf("database name is empty in the URL")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// We ignore many URI connection parameters and start pinger as soon as possible
	// because it takes a long time on CI to establish the Tailscale connection.

	ctx := context.Background()

	opts := options.Client().ApplyURI(uri)
	opts.SetDirect(true)
	opts.SetConnectTimeout(3 * time.Second)
	opts.SetHeartbeatInterval(3 * time.Second)
	opts.SetMaxConnIdleTime(0)
	opts.SetMinPoolSize(1)
	opts.SetMaxPoolSize(1)
	opts.SetMaxConnecting(1)

	l.InfoContext(ctx, "Connecting to MongoDB URI to push results...", slog.String("uri", u.Redacted()))

	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	pingerCtx, pingerCancel := context.WithCancel(ctx)

	res := &Client{
		l:            l,
		c:            c,
		pingerCancel: pingerCancel,
		pingerDone:   make(chan struct{}),
		database:     database,
		hostname:     hostname,
		runner:       os.Getenv("RUNNER_NAME"),
		repository:   os.Getenv("GITHUB_REPOSITORY"),
	}

	go func() {
		res.ping(pingerCtx)
		close(res.pingerDone)
	}()

	return res, nil
}

// ping pings the database until connection is established or ctx is canceled.
//
// TODO https://github.com/FerretDB/dance/issues/1122
func (c *Client) ping(ctx context.Context) {
	for ctx.Err() == nil {
		pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)

		err := c.c.Ping(pingCtx, nil)
		if err == nil {
			c.l.InfoContext(pingCtx, "Ping successful")
			pingCancel()

			return
		}

		c.l.WarnContext(pingCtx, "Ping failed", slog.String("error", err.Error()))

		// always wait, even if ping returns immediately
		<-pingCtx.Done()
		pingCancel()
	}
}

// Push pushes test results to MongoDB-compatible database.
func (c *Client) Push(ctx context.Context, config, database string, res map[string]config.TestResult) error {
	var passed bson.D

	for t, tr := range res {
		t = strings.ReplaceAll(t, ".", "_") // to make it compatible with FerretDB v1
		passed = append(passed, bson.E{t, bson.D{{"m", tr.Measurements}}})
	}

	doc := bson.D{
		{"config", config},
		{"database", database},
		{"time", time.Now()},
		{"env", bson.D{
			{"runner", c.runner},
			{"hostname", c.hostname},
			{"repository", c.repository},
		}},
		{"passed", passed},
	}

	c.l.InfoContext(ctx, "Pushing results to MongoDB URI...", slog.Any("doc", doc))

	c.ping(ctx)

	_, err := c.c.Database(c.database).Collection("incoming").InsertOne(ctx, doc)

	return err
}

// Close closes all connections.
//
// TODO https://github.com/FerretDB/dance/issues/1122
func (c *Client) Close() {
	c.pingerCancel()
	<-c.pingerDone

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.c.Disconnect(ctx)
}
