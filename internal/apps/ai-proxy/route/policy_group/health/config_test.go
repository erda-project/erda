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

package health

import (
	"testing"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
)

func TestNewManagerPanicWhenRescueBackoffUnset(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when rescue backoff is unset")
		}
	}()
	_ = NewManager(store, Config{
		Enabled: true,
		Probe: ProbeConfig{
			BaseURL:      "http://127.0.0.1:65530",
			UnhealthyTTL: time.Hour,
			Timeout:      20 * time.Millisecond,
		},
	})
}

func TestNewManagerDisabled(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{Enabled: false})
	if manager != nil {
		t.Fatal("expected nil manager when model health is disabled")
	}
}

func TestRetryUnhealthyFallbackWindow(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Probe: ProbeConfig{
			BaseURL:      "http://127.0.0.1:65530",
			UnhealthyTTL: time.Hour,
			Timeout:      20 * time.Millisecond,
		},
		Rescue: RescueConfig{
			InitialBackoff: 3 * time.Second,
			MaxBackoff:     2 * time.Minute,
		},
	}

	if got := RetryUnhealthyFallbackWindow(cfg); got != 10*time.Minute {
		t.Fatalf("expected retry unhealthy fallback window 10m, got %s", got)
	}
}
