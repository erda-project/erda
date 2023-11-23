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

package compressor

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func BenchmarkCompress(b *testing.B) {
	encoder := NewGzipEncoder(3)
	wait := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wait.Add(1)
		go func(idx int) {
			items := make([]*metric.Metric, 0, i)
			for j := 0; j < idx; j++ {
				items = append(items, &metric.Metric{
					Name:      "test",
					Timestamp: time.Now().UnixNano(),
					Tags: map[string]string{
						"tag1": RandStringRunes(rand.Intn(1000)),
						"tag2": RandStringRunes(rand.Intn(1000)),
					},
					Fields: map[string]interface{}{
						"field1": RandStringRunes(rand.Intn(1000)),
						"field2": RandStringRunes(rand.Intn(1000)),
					},
				})
			}
			obj := map[string][]*metric.Metric{
				"metrics": items,
			}
			buf, err := json.Marshal(&obj)
			if err != nil {
				b.Fatal(fmt.Errorf("serialize err: %w", err))
			}
			compressed, err := encoder.Compress(buf)
			assert.NoError(b, err)
			assert.NotEmpty(b, compressed)
			wait.Done()
		}(i)
	}
	wait.Wait()
}
