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
	"net/http"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) getOrgByRequest(r *http.Request) (*apistructs.OrgDTO, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return nil, errors.Errorf("missing org id header")
	}
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return nil, errors.Errorf("invalid org id")
	}
	org, err := e.bdl.GetOrg(orgID)
	if err != nil {
		return nil, err
	}
	return org, nil
}
