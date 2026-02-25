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

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "model", "series"), 0o755); err != nil {
		t.Fatalf("mkdir model dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "service_provider"), 0o755); err != nil {
		t.Fatalf("mkdir service_provider dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "model", "series", "a.json"), []byte(`{"model-a":{}}`), 0o644); err != nil {
		t.Fatalf("write model template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "service_provider", "sp.json"), []byte(`{"sp-a":{}}`), 0o644); err != nil {
		t.Fatalf("write service provider template: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"-path", tmp}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run code = %d, stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "template check passed") {
		t.Fatalf("stdout missing summary: %q", out)
	}
	if !strings.Contains(out, "service_provider=1") || !strings.Contains(out, "model=1") || !strings.Contains(out, "total=2") {
		t.Fatalf("stdout missing counts: %q", out)
	}
}

func TestRunFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-path", filepath.Join(t.TempDir(), "not-exists")}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "template check failed") {
		t.Fatalf("stderr missing failure message: %q", stderr.String())
	}
}
