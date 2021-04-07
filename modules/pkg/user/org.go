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
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

// GetOrgID 从 http request 的 header 中读取 org id.
func GetOrgID(r *http.Request) (uint64, error) {
	v := r.Header.Get("ORG-ID")

	orgID, err := strconv.ParseUint(v, 10, 64)
	if err == nil {
		return orgID, nil
	}

	return 0, errors.Errorf("invalid org id")
}
