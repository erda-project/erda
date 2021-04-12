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

package prehandle

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/pkg/httputil"
)

// FilterHeader remove internal header if exist in request
func FilterHeader(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	req.Header.Del(httputil.UserHeader)
	req.Header.Del(httputil.OrgHeader)
	req.Header.Del(httputil.InternalHeader)
	req.Header.Del(httputil.ClientIDHeader)
	req.Header.Del(httputil.ClientNameHeader)
}
