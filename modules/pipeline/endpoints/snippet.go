// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) querySnippetDetails(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	var req apistructs.SnippetQueryDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errorresp.ErrResp(err)
	}

	data, err := e.snippetSvc.QueryDetails(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data)
}
