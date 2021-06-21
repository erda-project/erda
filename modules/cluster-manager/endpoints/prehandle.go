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
	"net/http"

	"github.com/erda-project/erda/modules/cluster-manager/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type endpoint func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error)

func auth(f endpoint) endpoint {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
		// TODO: Use new auth
		if r.Header.Get(httputil.InternalHeader) == "" {
			return apierrors.ErrPreCheckCluster.AccessDenied().ToResp(), nil
		}

		if r.Body == nil {
			return apierrors.ErrPreCheckCluster.MissingParameter("body").ToResp(), nil
		}
		return f(ctx, r, vars)
	}
}
