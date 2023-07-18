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

package handlers

import (
	"net/http"
	"net/url"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type AuditsHandler struct {
	Log logs.Logger
	Dao dao.DAO
}

func (a *AuditsHandler) CreateAuditFromRequest(req *http.Request) error {
	panic("not implement")
}

func (a *AuditsHandler) UpdateFromResponse(id string, resp *http.Response) error {
	panic("not implement")
}

func (a *AuditsHandler) List(values url.Values) ([]map[string]any, error) {
	panic("not implement")
}
