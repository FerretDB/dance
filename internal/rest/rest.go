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

// Package rest contains RESTful API runner.
package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/FerretDB/dance/internal"
)

// Run runs RESTful API tests.
func Run(ctx context.Context, dir string, args []string, requests map[*http.Request]string) (*internal.TestResults, error) {
	// TODO https://github.com/FerretDB/dance/issues/20
	_ = ctx

	res := &internal.TestResults{
		TestResults: make(map[string]internal.TestResult),
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: config}

	client := &http.Client{Transport: tr}

	for r, payload := range requests {
		if payload != "" {
			payload = fmt.Sprintf("{%s}", payload)
			b, err := json.RawMessage([]byte(payload)).MarshalJSON()

			if err != nil {
				return nil, err
			}

			r.Body = io.NopCloser(bytes.NewBuffer(b))
		}

		req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
		if err != nil {
			return nil, err
		}
		req.Header = r.Header

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		out, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			res.TestResults[req.Method] = internal.TestResult{
				Status: internal.Pass,
				Output: string(out),
			}
		default:
			res.TestResults[req.Method] = internal.TestResult{
				Status: internal.Fail,
				Output: string(out),
			}
		}
	}

	return res, nil
}
