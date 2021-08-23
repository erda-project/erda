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
