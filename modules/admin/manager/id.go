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

package manager

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/http/httputil"
)

type USERID string

// Invalid return UserID is valid or invalid
func (uid USERID) Invalid() bool {
	return string(uid) == ""
}

func GetOrgID(req *http.Request) (uint64, error) {
	// get organization id
	orgIDStr := req.URL.Query().Get("orgID")
	if orgIDStr == "" {
		orgIDStr = req.Header.Get(httputil.OrgHeader)
		if orgIDStr == "" {
			return 0, errors.Errorf("invalid param, orgID is %s", orgIDStr)
		}
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return 0, errors.Errorf("invalid param, orgID is invalid")
	}
	return orgID, nil
}
