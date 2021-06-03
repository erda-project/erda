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

package fixedusers

import (
	"net/http"
	"strconv"

	"github.com/erda-project/erda/modules/openapi-ng/interceptors/auth/user"
)

type userInfo struct {
	id    string
	orgID uint64
}

func (u *userInfo) GetID() string {
	return u.id
}

func (u *userInfo) GetOrg() (uint64, bool) {
	if u.orgID == 0 {
		return 0, false
	}
	return u.orgID, true
}

func (u *userInfo) SetAuthInfo(r *http.Request) *http.Request {
	var orgID string
	if id, ok := u.GetOrg(); ok {
		orgID = strconv.FormatUint(id, 10)
	}
	return user.SetAuthInfo(u.GetID(), orgID, r)
}
