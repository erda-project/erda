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

package functions

import (
	"context"
	"encoding/json"
	"sync"
)

var (
	m  = make(map[string]FunctionFactory)
	mu = new(sync.Mutex)
)

type Function interface {
	Name() string
	Description() string
	SystemMessage() string
	UserMessage() string
	Schema() json.RawMessage
	Callback(ctx context.Context, arguments json.RawMessage) (any, error)
}

type Background struct {
	OrgID         uint64
	OrgName       string
	UserID        string
	Resources     map[string]json.Number
	ProjectID     uint64
	ApplicationID uint64
	Prompt        string `json:"prompt" yaml:"prompt"`
}

type FunctionFactory func(ctx context.Context, background *Background) Function

type CallbackURL struct{ CallbackURL any }

func Register(name string, factory FunctionFactory) {
	mu.Lock()
	m[name] = factory
	mu.Unlock()
}

func Retrieve(name string) (FunctionFactory, bool) {
	mu.Lock()
	ff, ok := m[name]
	mu.Unlock()
	return ff, ok
}
