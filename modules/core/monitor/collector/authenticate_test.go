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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/secret"
	"github.com/erda-project/erda/pkg/secret/hmac"
)

func TestAccessKeyAuthenticate(t *testing.T) {
	AccessKeyID, SecretKey := "ee3448nenk4B6efBxMBmT0Nr", "vJIC21Ze7U4Ofh65bz0K5475Y6O24bzu"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	coll := &provider{
		auth: &Authenticator{
			store: map[string]*model.AccessKey{
				AccessKeyID: {
					AccessKeyID: AccessKeyID,
					SecretKey:   SecretKey,
				},
			},
		},
		Cfg: &config{
			SignAuth: signAuthConfig{
				ExpiredDuration: time.Minute * 10,
			},
		},
	}
	handler := func(ctx httpserver.Context) error {
		return nil
	}

	h := coll.authSignedRequest()
	ass := assert.New(t)

	// Valid credentials
	keyPair := secret.AkSkPair{
		AccessKeyID: AccessKeyID,
		SecretKey:   SecretKey,
	}
	sign := hmac.New(keyPair, hmac.WithTimestamp(time.Now()))
	sign.SignCanonicalRequest(req)
	c := &mockContext{
		ctx: e.NewContext(req, res),
	}
	ass.NoError(h(handler)(c))

	// Expired Duration
	keyPair = secret.AkSkPair{
		AccessKeyID: AccessKeyID,
		SecretKey:   SecretKey,
	}
	sign = hmac.New(keyPair, hmac.WithTimestamp(time.Now().Add(-time.Minute*20)))
	sign.SignCanonicalRequest(req)
	c = &mockContext{
		ctx: e.NewContext(req, res),
	}
	ass.Error(h(handler)(c))

	// Invalid Secret Key
	keyPair = secret.AkSkPair{
		AccessKeyID: AccessKeyID,
		SecretKey:   "vJIC21Ze7U4Ofh65bz0K5475Y6O24xxx",
	}
	sign = hmac.New(keyPair, hmac.WithTimestamp(time.Now()))
	sign.SignCanonicalRequest(req)
	c = &mockContext{
		ctx: e.NewContext(req, res),
	}
	ass.Error(h(handler)(c))

	// Can't find accessKey
	keyPair = secret.AkSkPair{
		AccessKeyID: "xxxxxxxxx",
		SecretKey:   SecretKey,
	}
	sign = hmac.New(keyPair, hmac.WithTimestamp(time.Now()))
	sign.SignCanonicalRequest(req)
	c = &mockContext{
		ctx: e.NewContext(req, res),
	}
	ass.Error(h(handler)(c))

	// No Sign
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	c = &mockContext{
		ctx: e.NewContext(req, res),
	}
	ass.Error(h(handler)(c))
}

type mockContext struct {
	ctx echo.Context
}

func (m *mockContext) SetAttribute(key string, val interface{}) {
	return
}

func (m *mockContext) Attribute(key string) interface{} {
	return nil
}

func (m *mockContext) Attributes() map[string]interface{} {
	return nil
}

func (m *mockContext) Request() *http.Request {
	return m.ctx.Request()
}

func (m *mockContext) ResponseWriter() http.ResponseWriter {
	return nil
}

func (m *mockContext) Param(name string) string {
	return ""
}

func (m *mockContext) ParamNames() []string {
	return nil
}
