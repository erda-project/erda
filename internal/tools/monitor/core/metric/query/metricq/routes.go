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

package metricq

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/erda-project/erda-infra/providers/httpserver"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	api "github.com/erda-project/erda/pkg/common/httpapi"
	"github.com/erda-project/erda/pkg/discover"
)

func (p *provider) initRoutes(routes httpserver.Router) error {
	p.initRoutesV1(routes)
	// metric query apis
	routes.GET("/api/query", p.queryMetrics)  // for tsql
	routes.POST("/api/query", p.queryMetrics) // for tsql

	// Data export, temporary solution.
	return nil
}

// queryMetrics .
func (p *provider) queryMetrics(r *http.Request) interface{} {
	params := make(map[string]interface{})

	ctx := api.GetContext(r, func(header *http.Header) {
		if len(header.Get("org")) == 0 {
			//org-apis
			orgID := api.OrgID(r)
			if len(orgID) <= 0 {
				return
			}

			orgResp, err := p.Org.GetOrg(apis.WithInternalClientContext(context.Background(), discover.SvcMonitor), &orgpb.GetOrgRequest{
				IdOrName: orgID,
			})
			if err != nil {
				fmt.Println(err)
				return
			}
			header.Set("org", orgResp.Data.Name)
		}
	})

	err := r.ParseForm()
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	ql, q, format := r.Form.Get("ql"), r.Form.Get("q"), r.Form.Get("format")
	r.Form.Del("ql")
	r.Form.Del("q")
	r.Form.Del("format")
	if len(format) == 0 {
		format = "influxdb"
	}
	if len(ql) == 0 {
		ql = "influxql"
	}
	if len(q) == 0 {
		byts, err := io.ReadAll(r.Body)
		if err == nil {
			q = string(byts)
		}
	}
	resp, data, err := p.q.QueryWithFormat(ctx, ql, q, format, api.Language(r), params, nil, r.Form)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	if resp.Details != nil {
		return resp.Details
	}
	if response, ok := data.(httpserver.Response); ok {
		return response
	} else if response, ok := data.(httpserver.ResponseGetter); ok {
		return response
	}
	return api.Success(data)
}

func (p *provider) queryExternalMetrics(r *http.Request) interface{} {
	params := make(map[string]interface{})

	ctx := context.Background()

	err := r.ParseForm()
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	ql, q, format := r.Form.Get("ql"), r.Form.Get("q"), r.Form.Get("format")
	r.Form.Del("ql")
	r.Form.Del("q")
	r.Form.Del("format")
	if len(format) == 0 {
		format = "influxdb"
	}
	if len(ql) == 0 {
		ql = "influxql"
	}
	if len(q) == 0 {
		byts, err := io.ReadAll(r.Body)
		if err == nil {
			q = string(byts)
		}
	}
	resp, data, err := p.q.QueryExternalWithFormat(ctx, ql, q, format, api.Language(r), params, nil, r.Form)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	if resp.Details != nil {
		return resp.Details
	}
	if response, ok := data.(httpserver.Response); ok {
		return response
	} else if response, ok := data.(httpserver.ResponseGetter); ok {
		return response
	}
	return api.Success(data)
}
