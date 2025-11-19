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
