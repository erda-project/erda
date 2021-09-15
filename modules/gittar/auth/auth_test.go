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

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bou.ke/monkey"
	"github.com/labstack/echo"

	"github.com/erda-project/erda/modules/gittar/models"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

func TestIsGitProtocolRequest(t *testing.T) {
	e := echo.New()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("Git-Protocol", "version")
	res := httptest.NewRecorder()
	ctx1 := &webcontext.Context{
		EchoContext: e.NewContext(req1, res),
		Repository:  nil,
		User:        nil,
		Service:     nil,
		DBClient:    nil,
		Bundle:      nil,
		UCAuth:      nil,
		EtcdClient:  nil,
	}
	ctx2 := &webcontext.Context{
		EchoContext: e.NewContext(req2, res),
		Repository:  nil,
		User:        nil,
		Service:     nil,
		DBClient:    nil,
		Bundle:      nil,
		UCAuth:      nil,
		EtcdClient:  nil,
	}

	if isGitProtocolRequest(ctx1) != true {
		t.Error("fail")
	}
	if isGitProtocolRequest(ctx2) == true {
		t.Error("fail")
	}
}

func TestDoAuthWithHttpProtocolWithoutUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	ctx := &webcontext.Context{
		EchoContext: e.NewContext(req, res),
		Repository:  nil,
		User:        nil,
		Service:     nil,
		DBClient:    nil,
		Bundle:      nil,
		UCAuth:      nil,
		EtcdClient:  nil,
	}
	monkey.Patch(openRepository, func(ctx *webcontext.Context, repo *models.Repo) (*gitmodule.Repository, error) {
		return nil, nil
	})
	defer monkey.UnpatchAll()

	doAuth(ctx, nil, "")
	if ctx.EchoContext.Response().Status != 500 {
		t.Error("fail")
	}
}
