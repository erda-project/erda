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

package user

import (
	"net"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

// CheckOrg .
func CheckOrg(bdl *bundle.Bundle, r *http.Request, userID string) (bool, uint64, error) {
	orgHeader := r.Header.Get("ORG")
	var orgID uint64
	if orgHeader != "" && orgHeader != "-" {
		org, err := bdl.GetOrg(orgHeader)
		if err != nil {
			return false, 0, err
		}
		orgID = org.ID
	} else {
		domain, _, _ := net.SplitHostPort(r.Host)
		org, err := bdl.GetOrgByDomain(domain, userID)
		if err != nil {
			return false, 0, err
		}
		if org == nil {
			return true, 0, nil
		}
		orgID = org.ID
	}
	role, err := bdl.ScopeRoleAccess(userID, &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.OrgScope,
			ID:   strconv.FormatUint(orgID, 10),
		},
	})
	if err != nil {
		return false, 0, err
	}
	if !role.Access {
		return false, orgID, nil
	}
	return true, orgID, nil
}

// SetAuthInfo .
func SetAuthInfo(userID, orgID string, r *http.Request) *http.Request {
	r.Header.Set("User-ID", userID)
	if len(orgID) > 0 {
		r.Header.Set("Org-ID", orgID)
	}
	return r
}
