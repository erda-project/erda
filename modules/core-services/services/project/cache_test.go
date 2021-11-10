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

package project

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	memberC := NewCache(time.Millisecond * 200)
	quotaC := NewCache(time.Millisecond * 200)
	for i := 0; i < 50; i++ {
		member := new(memberCache)
		memberC.Store(i, &CacheItme{Object: member})
	}
	for i := 0; i < 20; i++ {
		quota := new(quotaCache)
		quotaC.Store(i, &CacheItme{Object: quota})
	}
	time.Sleep(time.Second)
	for i := 0; i < 20; i++ {
		value, ok := quotaC.Load(i)
		if !ok {
			t.Fatal("store error")
		}
		if isExpired := value.(*CacheItme).IsExpired(); !isExpired {
			t.Fatal("it should be expired")
		}

		value, ok = memberC.Load(i)
		if !ok {
			t.Fatal("store error")
		}
		if isExpired := value.(*CacheItme).IsExpired(); !isExpired {
			t.Fatal("it should be expired")
		}
	}
	quotaC.Store(1, &CacheItme{Object: new(quotaCache)})
	value, _ := quotaC.Load(1)
	if isExpired := value.(*CacheItme).IsExpired(); isExpired {
		t.Fatal("it should not be expired")
	}
}
