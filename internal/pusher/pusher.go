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
	l          *slog.Logger
	c          *mongo.Client
	database   string
	hostname   string
	runner     string
	repository string
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

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	l.InfoContext(ctx, "Connecting to MongoDB URI to push results...", slog.String("uri", u.Redacted()))

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err = c.Ping(ctx, nil); err != nil {
		c.Disconnect(ctx)
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &Client{
		l:          l,
		c:          c,
		database:   database,
		hostname:   hostname,
		runner:     os.Getenv("RUNNER_NAME"),
		repository: os.Getenv("GITHUB_REPOSITORY"),
	}, nil
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

	_, err := c.c.Database(c.database).Collection("incoming").InsertOne(ctx, doc)

	return err
}

// Close closes all connections.
func (c *Client) Close() {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	c.c.Disconnect(ctx)
}
