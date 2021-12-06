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
	"flag"
	"log"

	"github.com/FerretDB/dance/internal/gotest"
)

func main() {
	log.SetPrefix("dance: ")
	log.SetFlags(0)
	flag.Parse()

	res, err := gotest.Run(".")
	if err != nil {
		log.Fatal(err)
	}
	for t, res := range res.TestResults {
		log.Printf("%s: %+v", t, res)
	}

	// matches, err := filepath.Glob("*.yml")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for _, match := range matches {
	// 	log.Print(match)
	// }
}
