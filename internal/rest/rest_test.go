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

package rest

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	requests := map[*http.Request]string{
		// POST /api/games
		&http.Request{
			Method: http.MethodPost,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			URL: &url.URL{
				Scheme: "https",
				Host:   "localhost:5001",
				Path:   "/api/games",
			},
		}: `"Name":"The Legend of Zelda: Tears of the Kingdom", "Price": 57.95, "Category": "Action-adventure"`,
		// GET /api/games
		&http.Request{
			Method: http.MethodGet,
			URL: &url.URL{
				Scheme: "http",
				Host:   "localhost:5000",
				Path:   "/api/games/646bed57d305491677b25953",
			},
		}: "",
	}

	res, err := Run(ctx, "", nil, requests)
	assert.NoError(t, err)

	t.Log(res)
}
