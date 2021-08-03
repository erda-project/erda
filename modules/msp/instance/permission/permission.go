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

package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erda-project/erda-infra/providers/httpserver"
	httpperm "github.com/erda-project/erda/modules/monitor/common/permission"
	"github.com/erda-project/erda/modules/msp/instance"
	instancedb "github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/common/permission"
)

// Interface .
type Interface interface {
	TenantToProjectID(tgroup, tenantID string) permission.ValueGetter
	TerminusKeyToProjectID(key string) permission.ValueGetter
	TerminusKeyToProjectIDForHTTP(keys ...string) httpperm.ValueGetter
}

// TenantToProjectID .
func (p *provider) TenantToProjectID(tgroup, tenantID string) permission.ValueGetter {
	groupGetter := permission.FieldValue(tgroup)
	tenantGetter := permission.FieldValue(tenantID)
	return func(ctx context.Context, req interface{}) (string, error) {
		tg, _ := groupGetter(ctx, req)
		tID, _ := tenantGetter(ctx, req)
		idByTg, _ := p.getProjectIDByGroupIDOrTenantID(tg)
		if idByTg == "" {
			idByTID, err := p.getProjectIDByGroupIDOrTenantID(tID)
			if err != nil {
				return "", err
			}
			return idByTID, nil
		}
		return idByTg, nil
	}
}

func (p *provider) getProjectIDByGroupIDOrTenantID(id string) (string, error) {
	projectID, _ := p.getProjectIDByGroupID(id)
	if projectID == "" {
		return p.getProjectIDByTenantID(id)
	}
	return projectID, nil
}

func (p *provider) getProjectIDByGroupID(group string) (string, error) {
	id, err := p.getProjectIdByMSPTenantID(group)
	if err != nil {
		return "", errors.NewDatabaseError(err)
	}
	if id != "" {
		return id, nil
	}

	tenants, err := p.instanceTenantDB.GetByTenantGroup(group)
	if err != nil {
		return "", errors.NewDatabaseError(err)
	}
	if len(tenants) <= 0 {
		return "", errors.NewNotFoundError(group)
	}
	for _, tenant := range tenants {
		tmc, err := p.tmcDB.GetByEngine(tenant.Engine)
		if err != nil {
			return "", errors.NewDatabaseError(err)
		}
		if tmc == nil {
			continue
		}
		if strings.EqualFold(tmc.ServiceType, string(instance.ServiceTypeMicroService)) {
			id := p.getProjectIDByTenant(tenant)
			if len(id) > 0 {
				return id, nil
			}
		}
	}
	return "", errors.NewNotFoundError(group)
}

func (p *provider) getProjectIDByTenantID(id string) (string, error) {
	id, err := p.getProjectIdByMSPTenantID(id)
	if err != nil {
		return "", errors.NewDatabaseError(err)
	}
	if id != "" {
		return id, nil
	}

	tenants, err := p.instanceTenantDB.GetByTenantGroup(id)
	if err != nil {
		return "", errors.NewDatabaseError(err)
	}
	if len(tenants) <= 0 {
		return "", errors.NewNotFoundError(id)
	}
	for _, tenant := range tenants {
		tmc, err := p.tmcDB.GetByEngine(tenant.Engine)
		if err != nil {
			return "", errors.NewDatabaseError(err)
		}
		if tmc == nil {
			continue
		}
		if strings.EqualFold(tmc.ServiceType, string(instance.ServiceTypeMicroService)) {
			id := p.getProjectIDByTenant(tenant)
			if len(id) > 0 {
				return id, nil
			}
		}
	}
	return "", errors.NewNotFoundError(id)
}

func (p *provider) getProjectIdByMSPTenantID(id string) (string, error) {
	mspTenant, err := p.MSPTenantDB.QueryTenant(id)
	if err != nil {
		return "", err
	}
	if mspTenant != nil {
		return mspTenant.RelatedProjectId, nil
	}
	return "", nil
}

func (p *provider) getProjectIDByTenant(tenant *instancedb.InstanceTenant) string {
	if len(tenant.Options) <= 0 {
		return ""
	}
	options := make(map[string]interface{})
	json.Unmarshal([]byte(tenant.Options), &options)
	pid := options["projectId"]
	if pid == nil {
		return ""
	}
	return fmt.Sprint(pid)
}

func (p *provider) TerminusKeyToProjectID(terminusKey string) permission.ValueGetter {
	tkGetter := permission.FieldValue(terminusKey)
	return func(ctx context.Context, req interface{}) (string, error) {
		tk, err := tkGetter(ctx, req)
		if err != nil {
			return "", err
		}
		m, err := p.monitorDB.GetByTerminusKey(tk)
		if err != nil {
			return "", errors.NewDatabaseError(err)
		}
		return m.ProjectId, nil
	}
}

func (p *provider) TerminusKeyToProjectIDForHTTP(keys ...string) httpperm.ValueGetter {
	return func(ctx httpserver.Context) (string, error) {
		req := ctx.Request()
		for _, key := range keys {
			key := req.URL.Query().Get(key)
			if len(key) <= 0 {
				continue
			}
			id, err := p.getProjectIdByMSPTenantID(key)
			if err != nil {
				return "", errors.NewDatabaseError(err)
			}
			if id != "" {
				return id, nil
			}

			m, err := p.monitorDB.GetByTerminusKey(key)
			if err != nil {
				return "", fmt.Errorf("fail to get monitor: %s", err)
			}
			if m == nil || m.IsDelete > 0 {
				return "", fmt.Errorf("monitor not found")
			}
			return m.ProjectId, nil
		}
		return "", fmt.Errorf("terminus key not found")
	}
}
