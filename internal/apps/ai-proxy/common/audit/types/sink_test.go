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
	"reflect"
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


func TestNoteAppendBehavior(t *testing.T) {
	s := New("aid", logrusx.New())

	// first append: key not exist => set to val directly
	s.NoteAppend("arr.key", "v1")
	if got := s.Snapshot()["arr.key"]; got != "v1" {
		t.Fatalf("first NoteAppend should set to val, got %v", got)
	}

	// second append on string: concatenate
	s.NoteAppend("arr.key", "v2")
	if got := s.Snapshot()["arr.key"]; got != "v1v2" {
		t.Fatalf("second NoteAppend on string should concat, got %v", got)
	}

	// third append on string: still concatenate
	s.NoteAppend("arr.key", "v3")
	if got := s.Snapshot()["arr.key"]; got != "v1v2v3" {
		t.Fatalf("third NoteAppend on string should concat, got %v", got)
	}
}

func TestNoteAppendInvalidKeys(t *testing.T) {
	s := New("aid", logrusx.New())

	before := len(s.Snapshot())
	// empty key: should be ignored (no panic)
	s.NoteAppend("", 1)
	after := len(s.Snapshot())
	if after != before {
		t.Fatalf("empty key should be ignored, snapshot size changed: %d -> %d", before, after)
	}

	// invalid chars: contains space
	s.NoteAppend("bad key", 2)
	if len(s.Snapshot()) != before {
		t.Fatalf("invalid key should be ignored, snapshot size changed")
	}
}

func TestNoteAppendSliceAnyBulk(t *testing.T) {
	s := New("aid", logrusx.New())
	// preset []any
	s.Note("k", []any{"v1"})
	s.NoteAppend("k", "v2")
	got := s.Snapshot()["k"]
	slice, ok := got.([]any)
	if !ok {
		t.Fatalf("expected []any after append, got %T: %v", got, got)
	}
	if !reflect.DeepEqual(slice, []any{"v1", "v2"}) {
		t.Fatalf("want [v1 v2], got %v", slice)
	}
	// appending []any value should append as a single element (no flattening)
	s.NoteAppend("k", []any{"v3", "v4"})
	got = s.Snapshot()["k"]
	slice = got.([]any)
	if len(slice) != 3 {
		t.Fatalf("want len 3, got %d: %v", len(slice), slice)
	}
	if slice[0] != "v1" || slice[1] != "v2" {
		t.Fatalf("unexpected prefix: %v", slice)
	}
	if !reflect.DeepEqual(slice[2], []any{"v3", "v4"}) {
		t.Fatalf("the third element should be []any{\"v3\",\"v4\"}, got %v", slice[2])
	}
}

func TestNoteAppendSliceString_DefaultToAny(t *testing.T) {
	s := New("aid", logrusx.New())
	// preset []string
	s.Note("k", []string{"a"})
	s.NoteAppend("k", "b")
	got := s.Snapshot()["k"]
	anySlice, ok := got.([]any)
	if !ok {
		t.Fatalf("expected []any after appending to []string, got %T: %v", got, got)
	}
	if len(anySlice) != 2 || !reflect.DeepEqual(anySlice[0], []string{"a"}) || anySlice[1] != "b" {
		t.Fatalf("want [ [a] b ], got %v", anySlice)
	}
	// append []string next; since current is []any, v is appended as one element
	s.NoteAppend("k", []string{"c", "d"})
	got = s.Snapshot()["k"]
	anySlice, ok = got.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T: %v", got, got)
	}
	if len(anySlice) != 3 || !reflect.DeepEqual(anySlice[0], []string{"a"}) || anySlice[1] != "b" || !reflect.DeepEqual(anySlice[2], []string{"c", "d"}) {
		t.Fatalf("want [ [a] b [c d] ], got %v", anySlice)
	}
	// append mixed type -> simply appended into []any
	s.NoteAppend("k", 123)
	got = s.Snapshot()["k"]
	anySlice, ok = got.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T: %v", got, got)
	}
	if len(anySlice) != 4 || !reflect.DeepEqual(anySlice[0], []string{"a"}) || anySlice[1] != "b" || !reflect.DeepEqual(anySlice[2], []string{"c", "d"}) || anySlice[3] != 123 {
		t.Fatalf("unexpected slice: %v", anySlice)
	}
}

func TestNoteAppendBytes_Simple(t *testing.T) {
	s := New("aid", logrusx.New())
	s.Note("kb", []byte("ab"))
	s.NoteAppend("kb", []byte("cd"))
	got := s.Snapshot()["kb"]
	bs, ok := got.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T: %v", got, got)
	}
	if string(bs) != "abcd" {
		t.Fatalf("want abcd, got %q", string(bs))
	}
}
