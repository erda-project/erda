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

package reconciler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/throttler"
	"github.com/erda-project/erda/pkg/jsonstore"
)

type mockThrottler struct{}

func (m *mockThrottler) Name() string                                                     { return "" }
func (m *mockThrottler) AddQueue(name string, window int64)                               { return }
func (m *mockThrottler) AddKeyToQueues(key string, reqs []throttler.AddKeyToQueueRequest) { return }
func (m *mockThrottler) PopPending(key string) (bool, []throttler.PopDetail)              { return true, nil }
func (m *mockThrottler) PopProcessing(key string) (bool, []throttler.PopDetail)           { return true, nil }
func (m *mockThrottler) Export() json.RawMessage                                          { return nil }
func (m *mockThrottler) Import(message json.RawMessage) error                             { return nil }

type js struct{ jsonstore.JsonStoreImpl }

func (j *js) Put(ctx context.Context, key string, object interface{}) error {
	time.Sleep(time.Second * 2) // larger than ctx timeout
	return nil
}

func TestContinueBackupThrottler(t *testing.T) {
	th := &mockThrottler{}
	j := &js{}
	r := &Reconciler{TaskThrottler: th, js: j}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.continueBackupThrottler(ctx)
	// non blocking
}
