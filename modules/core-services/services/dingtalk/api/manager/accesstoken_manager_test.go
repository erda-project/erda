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

package manager

import (
	"sync"
	"testing"
)

func Test_GetAccessTokenManager_WithMultipleTimes_Should_CreateOnlyOneRequestLock(t *testing.T) {
	m := &Manager{}
	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			tm := m.GetAccessTokenManager("mock_appkey", "mock_appsecret")
			if tm == nil {
				t.Errorf("GetAccessTokenManager should not return nil")
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if len(requestLocks) != 1 {
		t.Errorf("concurrency get token manager for same key, should create only one lock")
	}
}
