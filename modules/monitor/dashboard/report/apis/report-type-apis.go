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

package apis

import (
	"net/http"

	"github.com/erda-project/erda-infra/modcom/api"
	dicestructs "github.com/erda-project/erda/apistructs"
)

func (p *provider) listReportType(r *http.Request, params struct {
	Scope string `param:"scope"`
}) interface{} {
	if params.Scope == "" {
		params.Scope = dicestructs.OrgResource
	}
	types := []reportType{
		{Name: p.t.Text(api.Language(r), "日报"), Value: daily},
		{Name: p.t.Text(api.Language(r), "周报"), Value: weekly},
	}
	resp := reportTypeResp{
		Types: types,
		Total: len(types),
	}
	return api.Success(resp)
}
