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

package collector

import (
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func Test_provider_Init(t *testing.T) {
	type fields struct {
		Router httpserver.Router
	}
	type args struct {
		ctx servicehub.Context
	}

	p := &provider{
		Router: &mockRouter{t: t},
		Cfg:    &config{},
	}
	p.Cfg.Auth.Skip = true
	assert.Nil(t, p.Init(nil))
}

type mockRouter struct {
	t *testing.T
}

func (m *mockRouter) GET(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) POST(path string, handler interface{}, options ...interface{}) {
	assert.IsType(m.t, echo.MiddlewareFunc(nil), options[0])
	return
}

func (m *mockRouter) DELETE(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) PUT(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) PATCH(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) HEAD(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) CONNECT(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) OPTIONS(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) TRACE(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) Any(path string, handler interface{}, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) Static(prefix, root string, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) File(path, filepath string, options ...interface{}) {
	panic("implement me")
}

func (m *mockRouter) Add(method, path string, handler interface{}, options ...interface{}) error {
	panic("implement me")
}
