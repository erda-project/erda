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

package legacycontainer

import (
	"testing"
)

type testSingletonA struct {
	Value int
}

type testSingletonB struct {
	Name string
}

func TestRegister_Get(t *testing.T) {
	a := &testSingletonA{Value: 42}
	Register(a)
	defer func() {
	}()

	got := Get[*testSingletonA]()
	if got != a {
		t.Errorf("Get[*testSingletonA]() = %p, want %p", got, a)
	}
	if got.Value != 42 {
		t.Errorf("Get[*testSingletonA]().Value = %d, want 42", got.Value)
	}
}

func TestGet_notRegistered(t *testing.T) {
	got := Get[*testSingletonB]()
	if got != nil {
		t.Errorf("Get[*testSingletonB]() without register should return nil, got %v", got)
	}
}

func TestRegister_twicePanics(t *testing.T) {
	type duplicateType struct{}
	Register(&duplicateType{})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Register same type twice should panic")
		}
		if _, ok := r.(string); !ok {
			t.Errorf("panic message should be string, got %T", r)
		}
	}()
	Register(&duplicateType{})
}

func TestRegister_Get_valueType(t *testing.T) {
	type valueType struct {
		X int
	}
	v := valueType{X: 1}
	Register(v)

	got := Get[valueType]()
	if got.X != 1 {
		t.Errorf("Get[valueType]().X = %d, want 1", got.X)
	}
}

func TestGet_interfaceType(t *testing.T) {
	type iface interface {
		M() int
	}
	got := Get[iface]()
	if got != nil {
		t.Errorf("Get[iface]() without register should return nil, got %v", got)
	}
}
