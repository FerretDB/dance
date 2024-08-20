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

package runner

import (
	"bytes"
	"sync"
)

// LockedBuffer is a thread-safe [bytes.Buffer].
type LockedBuffer struct {
	m sync.Mutex
	b bytes.Buffer
}

// Bytes calls [bytes.Buffer.Write] in a thread-safe way.
func (lb *LockedBuffer) Write(p []byte) (int, error) {
	lb.m.Lock()
	defer lb.m.Unlock()

	return lb.b.Write(p)
}

// Bytes calls [bytes.Buffer.Bytes] in a thread-safe way.
func (lb *LockedBuffer) Bytes() []byte {
	lb.m.Lock()
	defer lb.m.Unlock()

	return lb.b.Bytes()
}
