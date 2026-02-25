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

package template

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
)

func TestLoadTemplatesFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"model/series/a.json":      {Data: []byte(`{"model-a":{},"model-b":{}}`)},
		"service_provider/sp.json": {Data: []byte(`{"sp-a":{}}`)},
		"model/series/readme.txt":  {Data: []byte("ignored")},
	}

	templatesByType, err := LoadTemplatesFromFS(nil, fsys)
	if err != nil {
		t.Fatalf("LoadTemplatesFromFS returned error: %v", err)
	}

	if got := len(templatesByType[templatetypes.TemplateTypeServiceProvider]); got != 1 {
		t.Fatalf("service_provider count = %d, want 1", got)
	}
	if got := len(templatesByType[templatetypes.TemplateTypeModel]); got != 2 {
		t.Fatalf("model count = %d, want 2", got)
	}

	summary := TemplateCheckSummary(templatesByType)
	if !strings.Contains(summary, "service_provider=1") || !strings.Contains(summary, "model=2") || !strings.Contains(summary, "total=3") {
		t.Fatalf("unexpected summary: %s", summary)
	}
}

func TestLoadTemplatesFromFSDuplicateName(t *testing.T) {
	fsys := fstest.MapFS{
		"model/series/a.json": {Data: []byte(`{"dup":{}}`)},
		"model/series/b.json": {Data: []byte(`{"dup":{}}`)},
	}

	_, err := LoadTemplatesFromFS(nil, fsys)
	if err == nil {
		t.Fatalf("expected duplicate template error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate template") {
		t.Fatalf("unexpected error: %v", err)
	}
}
