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

package google_vertex_ai_director

import (
	"context"
	"sync"
	"testing"

	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
)

func resetGCPProjectIDCache() {
	gcpProjectIDCache = sync.Map{}
}

func TestGetGCPProjectIdCachesParsedValue(t *testing.T) {
	resetGCPProjectIDCache()

	content := `{"project_id":"proj-123"}`
	sp := &providerpb.ServiceProvider{
		TemplateParams: map[string]string{
			"service-account-key-file-content": content,
		},
	}

	pid, err := getGCPProjectId(context.Background(), sp)
	if err != nil {
		t.Fatalf("getGCPProjectId returned error: %v", err)
	}
	if pid != "proj-123" {
		t.Fatalf("expected project_id proj-123, got %s", pid)
	}

	cacheKey := cacheKeyFromContent(content)
	if v, ok := gcpProjectIDCache.Load(cacheKey); !ok || v.(string) != "proj-123" {
		t.Fatalf("expected cache to store project_id under hashed key")
	}
}

func TestGetGCPProjectIdUsesHashedCacheKey(t *testing.T) {
	resetGCPProjectIDCache()

	content := `invalid-json`
	cacheKey := cacheKeyFromContent(content)
	gcpProjectIDCache.Store(cacheKey, "cached-proj")

	sp := &providerpb.ServiceProvider{
		TemplateParams: map[string]string{
			"service-account-key-file-content": content,
		},
	}

	pid, err := getGCPProjectId(context.Background(), sp)
	if err != nil {
		t.Fatalf("expected cached project id, got error: %v", err)
	}
	if pid != "cached-proj" {
		t.Fatalf("expected cached project id cached-proj, got %s", pid)
	}
}
