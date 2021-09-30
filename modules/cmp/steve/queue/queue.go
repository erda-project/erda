// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queue

import (
	"os"
	"strconv"
)

var queryQueue chan struct{}
var queueSize int

func init() {
	queueSize = 5
	if size, err := strconv.Atoi(os.Getenv("STEVE_QUEUE_SIZE")); err == nil && size > queueSize {
		queueSize = size
	}
	queryQueue = make(chan struct{}, queueSize)
}

func Acquire(delta int) {
	for i := 0; i < delta; i++ {
		queryQueue <- struct{}{}
	}
}

func Release(delta int) {
	for i := 0; i < delta; i++ {
		<-queryQueue
	}
}

func Length() int {
	return queueSize
}
