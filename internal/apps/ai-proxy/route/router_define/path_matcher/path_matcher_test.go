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

package path_matcher

import (
	"sync"
	"testing"
)

func TestNewPathMatcher(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
		params  map[string]string
	}{
		{
			name:    "exact match without parameters",
			pattern: "/api/v1/models",
			path:    "/api/v1/models",
			want:    true,
			params:  map[string]string{},
		},
		{
			name:    "no match different path",
			pattern: "/api/v1/models",
			path:    "/api/v1/users",
			want:    false,
			params:  map[string]string{},
		},
		{
			name:    "single parameter match",
			pattern: "/api/files/{file_id}",
			path:    "/api/files/123",
			want:    true,
			params:  map[string]string{"file_id": "123"},
		},
		{
			name:    "multiple parameters match",
			pattern: "/api/files/{file_id}/versions/{version_id}",
			path:    "/api/files/123/versions/456",
			want:    true,
			params:  map[string]string{"file_id": "123", "version_id": "456"},
		},
		{
			name:    "parameter with complex value",
			pattern: "/api/models/{model_name}",
			path:    "/api/models/gpt-4o-mini",
			want:    true,
			params:  map[string]string{"model_name": "gpt-4o-mini"},
		},
		{
			name:    "parameter should not match across slashes",
			pattern: "/api/files/{file_id}",
			path:    "/api/files/123/extra",
			want:    false,
			params:  map[string]string{},
		},
		{
			name:    "special characters in path",
			pattern: "/api/files/{file_id}",
			path:    "/api/files/file-123_test.txt",
			want:    true,
			params:  map[string]string{"file_id": "file-123_test.txt"},
		},
		{
			name:    "empty parameter value should not match",
			pattern: "/api/files/{file_id}",
			path:    "/api/files/",
			want:    false,
			params:  map[string]string{},
		},
		{
			name:    "pattern with dots and special chars",
			pattern: "/api/v1.0/files/{file_id}",
			path:    "/api/v1.0/files/123",
			want:    true,
			params:  map[string]string{"file_id": "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPathMatcher(tt.pattern)

			if pm == nil {
				t.Fatal("NewPathMatcher returned nil")
			}

			got := pm.Match(tt.path)
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}

			if tt.want {
				for key, expectedValue := range tt.params {
					actualValue, exists := pm.GetValue(key)
					if !exists {
						t.Errorf("Parameter %s not found in Values", key)
					}
					if actualValue != expectedValue {
						t.Errorf("Parameter %s = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestPathMatcher_SetValue(t *testing.T) {
	pm := NewPathMatcher("/test")

	pm.SetValue("key1", "value1")
	pm.SetValue("key2", "value2")

	value1, exists1 := pm.GetValue("key1")
	if !exists1 || value1 != "value1" {
		t.Errorf("GetValue(key1) = %v, %v, want value1, true", value1, exists1)
	}

	value2, exists2 := pm.GetValue("key2")
	if !exists2 || value2 != "value2" {
		t.Errorf("GetValue(key2) = %v, %v, want value2, true", value2, exists2)
	}
}

func TestPathMatcher_GetValue_NotExists(t *testing.T) {
	pm := NewPathMatcher("/test")

	value, exists := pm.GetValue("nonexistent")
	if exists || value != "" {
		t.Errorf("GetValue(nonexistent) = %v, %v, want empty string, false", value, exists)
	}
}

func TestPathMatcher_ThreadSafety(t *testing.T) {
	pm := NewPathMatcher("/api/files/{file_id}")

	var wg sync.WaitGroup
	iterations := 100

	// Test concurrent SetValue calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pm.SetValue("key", "value")
		}(i)
	}

	// Test concurrent GetValue calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pm.GetValue("key")
		}()
	}

	// Test concurrent Match calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pm.Match("/api/files/123")
		}()
	}

	wg.Wait()

	// Verify the matcher still works correctly
	if !pm.Match("/api/files/test123") {
		t.Error("PathMatcher failed after concurrent operations")
	}

	value, exists := pm.GetValue("file_id")
	if !exists || value != "test123" {
		t.Errorf("GetValue after concurrent operations = %v, %v, want test123, true", value, exists)
	}
}

func TestPathMatcher_OverwriteValues(t *testing.T) {
	pm := NewPathMatcher("/api/files/{file_id}")

	// First match
	pm.Match("/api/files/123")
	value1, _ := pm.GetValue("file_id")
	if value1 != "123" {
		t.Errorf("First match: file_id = %v, want 123", value1)
	}

	// Second match should overwrite
	pm.Match("/api/files/456")
	value2, _ := pm.GetValue("file_id")
	if value2 != "456" {
		t.Errorf("Second match: file_id = %v, want 456", value2)
	}
}

func TestPathMatcher_ComplexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
		params  map[string]string
	}{
		{
			name:    "nested parameters",
			pattern: "/api/{version}/files/{file_id}/metadata/{meta_key}",
			path:    "/api/v1/files/123/metadata/size",
			want:    true,
			params:  map[string]string{"version": "v1", "file_id": "123", "meta_key": "size"},
		},
		{
			name:    "parameter at beginning",
			pattern: "/{version}/api/files",
			path:    "/v1/api/files",
			want:    true,
			params:  map[string]string{"version": "v1"},
		},
		{
			name:    "parameter at end",
			pattern: "/api/files/{file_id}",
			path:    "/api/files/final_file",
			want:    true,
			params:  map[string]string{"file_id": "final_file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPathMatcher(tt.pattern)
			got := pm.Match(tt.path)

			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}

			if tt.want {
				for key, expectedValue := range tt.params {
					actualValue, exists := pm.GetValue(key)
					if !exists {
						t.Errorf("Parameter %s not found", key)
					}
					if actualValue != expectedValue {
						t.Errorf("Parameter %s = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}
