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

package tenant

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type Tenant struct {
	db *dbclient.DBClient
	hc *httpclient.HTTPClient
}

// Option 应用实例对象配置选项
type Option func(*Tenant)

// New 新建应用实例 service
func New(options ...Option) *Tenant {
	t := &Tenant{}
	for _, op := range options {
		op(t)
	}
	return t
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(r *Tenant) {
		r.db = db
	}
}

// WithHTTPClient 配置 http 客户端对象.
func WithHTTPClient(hc *httpclient.HTTPClient) Option {
	return func(a *Tenant) {
		a.hc = hc
	}
}

func (t *Tenant) GetTenant(projectID int64, workspace, tenantGroup, userId string) (string, error) {
	monitor, err := t.db.GetMonitorByProjectIdAndWorkspace(projectID, workspace)
	if err != nil {
		return "", err
	}
	if monitor != nil {
		return tenantGroup, nil
	}

	var resp map[string]interface{}
	req := pb.CreateTenantRequest{
		ProjectID:  projectID,
		Workspaces: []string{workspace},
		TenantType: pb.Type_DOP.String(),
	}

	r, err := t.hc.Get(discover.MSP()).
		Path("/api/msp/tenant").
		Header("USER-ID", userId).
		JSONBody(&req).
		Do().
		JSON(&resp)
	var response pb.CreateTenantResponse
	if err != nil {
		return "", nil
	}
	err = json.Unmarshal(r.Body(), &response)
	if err != nil {
		return "", nil
	}
	if !r.IsOK() || len(response.Data) < 1 {
		logrus.Errorf("provider response err : %+v", r)
		return "", err
	}
	return response.Data[0].Id, nil
}
