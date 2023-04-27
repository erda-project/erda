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

package filter

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

const (
	Continue Signal = iota
	Intercept
)

var (
	instFuncs = make(map[string]InstantiateFunc)
	l         = new(sync.Mutex)
)

type Filter interface{}

type RequestGetterFilter interface {
	OnHttpRequestGetter(ctx context.Context, g HttpInfor) (Signal, error)
}

type RequestFilter interface {
	OnHttpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (Signal, error)
}

type ResponseGetterFilter interface {
	OnHttpResponseGetter(ctx context.Context, g HttpInfor) (Signal, error)
}

type ResponseFilter interface {
	OnHttpResponse(ctx context.Context, response *http.Response) (Signal, error)
}

type Config struct {
	Name   string          `json:"name" yaml:"name"`
	Config json.RawMessage `json:"config" yaml:"config"`
}

type Signal int

type InstantiateFunc func(config json.RawMessage) (Filter, error)

func Register(name string, inst InstantiateFunc) {
	l.Lock()
	defer l.Unlock()
	instFuncs[name] = inst
}

func Deregister(name string) {
	l.Lock()
	defer l.Unlock()
	delete(instFuncs, name)
}

func GetFilterFactory(name string) (InstantiateFunc, bool) {
	f, ok := instFuncs[name]
	return f, ok
}
