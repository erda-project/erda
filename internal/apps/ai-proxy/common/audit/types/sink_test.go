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

package types

import (
	"context"
	"testing"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
)

type mockWriter struct {
	calls int
	last  Patch
}

func (m *mockWriter) Write(ctx context.Context, p Patch) {
	m.calls++
	m.last = p
}

func TestNoteAndFlushBasic(t *testing.T) {
	s := New("aid-123", logrusx.New())
	s.Note("status", 200)
	s.Note("usage.prompt_tokens", 12)

	mw := &mockWriter{}
	s.Flush(context.Background(), mw)
	if mw.calls != 1 {
		t.Fatalf("expected 1 write, got %d", mw.calls)
	}
	if mw.last.AuditID != "aid-123" {
		t.Fatalf("unexpected audit id: %s", mw.last.AuditID)
	}
	if got := mw.last.Notes["status"]; got != 200 {
		t.Fatalf("want status=200, got %v", got)
	}
	if got := mw.last.Notes["usage.prompt_tokens"]; got != 12 {
		t.Fatalf("want prompt=12, got %v", got)
	}
}

func TestNoteErrors(t *testing.T) {
	// Note: errors are now logged as warnings instead of returned
	// This test verifies the methods don't panic
	s := New("aid", logrusx.New())
	s.Note("", 1) // should log warning
	s.NoteOnce("BadUpper", 1) // should log warning and return false
}

func TestNoteOnceSemantics(t *testing.T) {
	s := New("aid", logrusx.New())
	wrote := s.NoteOnce("route.director", "openai")
	if !wrote {
		t.Fatalf("first NoteOnce should write, wrote=%v", wrote)
	}
	wrote = s.NoteOnce("route.director", "anthropic")
	if wrote {
		t.Fatalf("second NoteOnce should not write, wrote=%v", wrote)
	}
	// Note should overwrite existing value
	s.Note("route.director", "azure")
	mw := &mockWriter{}
	s.Flush(context.Background(), mw)
	if got := mw.last.Notes["route.director"]; got != "azure" {
		t.Fatalf("want director=azure, got %v", got)
	}
}

func TestIncBehavior(t *testing.T) {
	s := New("aid", logrusx.New())
	// start from empty
	n := s.Inc("usage.completion_tokens", 5)
	if n != 5 {
		t.Fatalf("inc from empty want 5, got %d", n)
	}
	n = s.Inc("usage.completion_tokens", 3)
	if n != 8 {
		t.Fatalf("inc second want 8, got %d", n)
	}
	// non-numeric existing resets to 0 before inc
	s.Note("misc", "abc")
	n = s.Inc("misc", 2)
	if n != 2 {
		t.Fatalf("inc after non-numeric want 2, got %d", n)
	}
}

func TestSnapshotIsolation(t *testing.T) {
	s := New("aid", logrusx.New())
	s.Note("k", "v")
	snap := s.Snapshot()
	snap["k"] = "mutated"
	mw := &mockWriter{}
	s.Flush(context.Background(), mw)
	if got := mw.last.Notes["k"]; got != "v" {
		t.Fatalf("snapshot should be a copy; got %v", got)
	}
}
