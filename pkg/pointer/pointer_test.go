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

package pointer

import (
	"testing"
	"time"
)

func TestTo(t *testing.T) {
	p := To(42)
	if p == nil || *p != 42 {
		t.Errorf("To(42) = %v, want *42", p)
	}

	s := To("hello")
	if s == nil || *s != "hello" {
		t.Errorf("To(\"hello\") = %v, want *\"hello\"", s)
	}

	b := To(true)
	if b == nil || !*b {
		t.Errorf("To(true) = %v, want *true", b)
	}
}

func TestDeref(t *testing.T) {
	if got := Deref[int](nil, 0); got != 0 {
		t.Errorf("Deref(nil, 0) = %d, want 0", got)
	}
	if got := Deref(To(100), 0); got != 100 {
		t.Errorf("Deref(To(100), 0) = %d, want 100", got)
	}
	if got := Deref((*string)(nil), "default"); got != "default" {
		t.Errorf("Deref(nil, \"default\") = %q, want \"default\"", got)
	}
	if got := Deref(To("x"), "default"); got != "x" {
		t.Errorf("Deref(To(\"x\"), \"default\") = %q, want \"x\"", got)
	}
}

func TestDerefPtr(t *testing.T) {
	def := 99
	got := DerefPtr(nil, def)
	if got == nil || *got != 99 {
		t.Errorf("DerefPtr(nil, 99) = %v, want *99", got)
	}

	p := To(1)
	got2 := DerefPtr(p, 0)
	if got2 != p || *got2 != 1 {
		t.Errorf("DerefPtr(To(1), 0) = %v, want same ptr with 1", got2)
	}
}

func TestMustDeref(t *testing.T) {
	if got := MustDeref(To(7)); got != 7 {
		t.Errorf("MustDeref(To(7)) = %d, want 7", got)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustDeref(nil) should panic")
		}
		if r != "pointer: MustDeref nil pointer" {
			t.Errorf("panic message = %q, want %q", r, "pointer: MustDeref nil pointer")
		}
	}()
	MustDeref[int](nil)
}

func TestConvenienceVars(t *testing.T) {
	if p := String("s"); p == nil || *p != "s" {
		t.Errorf("String(\"s\") = %v", p)
	}
	if p := Int(1); p == nil || *p != 1 {
		t.Errorf("Int(1) = %v", p)
	}
	if p := Int64(2); p == nil || *p != 2 {
		t.Errorf("Int64(2) = %v", p)
	}
	if p := Uint8(3); p == nil || *p != 3 {
		t.Errorf("Uint8(3) = %v", p)
	}
	if p := Bool(true); p == nil || !*p {
		t.Errorf("Bool(true) = %v", p)
	}

	now := time.Now()
	if p := Time(now); p == nil || !p.Equal(now) {
		t.Errorf("Time(now) = %v", p)
	}

	if got := StringDeref(nil, "def"); got != "def" {
		t.Errorf("StringDeref(nil, \"def\") = %q", got)
	}
	if got := IntDeref(To(10), 0); got != 10 {
		t.Errorf("IntDeref(To(10), 0) = %d", got)
	}
	if got := Int64Deref(nil, 100); got != 100 {
		t.Errorf("Int64Deref(nil, 100) = %d", got)
	}
	if got := BoolDeref(To(false), true); got != false {
		t.Errorf("BoolDeref(To(false), true) = %v", got)
	}

	tt := time.Unix(1, 0)
	if got := TimeDeref(To(tt), time.Time{}); !got.Equal(tt) {
		t.Errorf("TimeDeref(To(tt), zero) = %v", got)
	}
	if got := TimeDeref(nil, tt); !got.Equal(tt) {
		t.Errorf("TimeDeref(nil, tt) = %v", got)
	}
}
