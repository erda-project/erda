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

// Package types provides a minimal audit Sink implementation.
// This version uses a plain map and assumes per-request serial access
// (filters in the same request run in a single goroutine chain).
package types

import (
	"context"
	"reflect"
	"regexp"

	"github.com/erda-project/erda-infra/base/logs"
)

// Sink is the minimal interface for recording audit notes.
// All methods are intended to be called serially for a single request.
type Sink interface {
	// Note sets/overwrites a key with val.
	// Prints warning if key is empty or invalid.
	Note(key string, val any)
	// NoteOnce sets key with val only if it hasn't been written before.
	// It is a no-op if the key already exists or NoteOnce was called earlier.
	// Returns true if the key was set, false if it already existed.
	NoteOnce(key string, val any) bool
	// Inc increments a numeric key by delta and returns the new value.
	// If the key doesn't exist or isn't numeric, it starts from 0.
	Inc(key string, delta int64) int64
	// Flush snapshots current notes and writes them via Writer.
	Flush(ctx context.Context, w Writer)
	// Snapshot makes a copy of current notes (for debugging/testing).
	Snapshot() map[string]any
}

// Writer persists a Patch assembled from a Sink.
type Writer interface {
	Write(ctx context.Context, p Patch)
}

// Patch is the unit written by Writer.
type Patch struct {
	AuditID string
	Notes   map[string]any // flat KV, e.g. "usage.prompt_tokens": 123
}

// New constructs a new Sink for the given auditID.
func New(auditID string, logger logs.Logger) Sink {
	return &sink{auditID: auditID, notes: make(map[string]any), onceSet: make(map[string]struct{}), logger: logger}
}

// sink is a plain-map implementation intended for per-request serial use.
type sink struct {
	auditID string
	notes   map[string]any      // k -> value
	onceSet map[string]struct{} // tracks NoteOnce-written keys

	logger logs.Logger
}

var (
	keyRe = regexp.MustCompile(`^[a-z0-9_.-]{1,128}$`)
)

func (s *sink) Note(k string, v any) {
	if k == "" {
		s.logger.Warnf("note key is empty, ignoring")
		return
	}
	if !keyRe.MatchString(k) {
		s.logger.Warnf("note key invalid: %s, allowed [a-z0-9_.-], 1..128", k)
		return
	}
	s.notes[k] = v
}

func (s *sink) NoteOnce(k string, v any) bool {
	if k == "" {
		s.logger.Warnf("note key is empty, ignoring")
		return false
	}
	if !keyRe.MatchString(k) {
		s.logger.Warnf("note key invalid: %s, allowed [a-z0-9_.-], 1..128", k)
		return false
	}
	// If already set (by Note or previous NoteOnce), do nothing.
	if _, exists := s.notes[k]; exists {
		return false
	}
	if _, exists := s.onceSet[k]; exists {
		return false
	}
	s.onceSet[k] = struct{}{}
	s.notes[k] = v
	return true
}

func (s *sink) Inc(k string, d int64) int64 {
	if k == "" {
		s.logger.Warnf("note key is empty, ignoring")
		return 0
	}
	if !keyRe.MatchString(k) {
		s.logger.Warnf("note key invalid: %s, allowed [a-z0-9_.-], 1..128", k)
		return 0
	}
	var cur int64
	switch v := s.notes[k].(type) {
	case int, int8, int16, int32, int64:
		// use reflection-free conversion via interface{}
		cur = reflect.ValueOf(v).Int()
	case float32, float64:
		cur = int64(reflect.ValueOf(v).Float())
	default:
		s.logger.Warnf("note key %s is not numeric (%v), starting from 0", k, v)
		cur = 0
	}
	n := cur + d
	s.notes[k] = n
	return n
}

func (s *sink) Snapshot() map[string]any {
	out := make(map[string]any, len(s.notes))
	for k, v := range s.notes {
		out[k] = v
	}
	return out
}

func (s *sink) Flush(ctx context.Context, w Writer) {
	w.Write(ctx, Patch{
		AuditID: s.auditID,
		Notes:   s.Snapshot(),
	})
}
