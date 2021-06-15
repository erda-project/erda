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

	"github.com/erda-project/erda/modules/msp/instance"
	instancedb "github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/pkg/common/permission"
)

// Interface .
type Interface interface {
	TenantToProjectID(tgroup, tenant string) permission.ValueGetter
}

// TenantToProjectID .
func (p *provider) TenantToProjectID(tgroup, tenant string) permission.ValueGetter {
	groupGetter := permission.FieldValue(tgroup)
	tenantGetter := permission.FieldValue(tenant)
	return func(ctx context.Context, req interface{}) (string, error) {
		if len(tgroup) > 0 {
			group, err := groupGetter(ctx, req)
			if err == nil && len(group) > 0 {
				return p.getProjectIDByGroupID(group)
			}
		}
		id, err := tenantGetter(ctx, req)
		if err != nil {
			return "", err
		}
		return p.getProjectIDByTenantID(id)
	}
}

func (p *provider) getProjectIDByTenantID(id string) (string, error) {
	tenant, err := p.instanceTenantDB.GetByID(id)
	if err != nil {
		return "", err
	}
	if tenant == nil {
		return "", fmt.Errorf("fail to find tenant by id %q", id)
	}

	return p.getProjectIDByTenant(tenant)
}

func (p *provider) getProjectIDByGroupID(group string) (string, error) {
	tenants, err := p.instanceTenantDB.GetByTenantGroup(group)
	if err != nil {
		return "", err
	}
	if len(tenants) <= 0 {
		return "", fmt.Errorf("tenant group %q not found", group)
	}
	for _, tenant := range tenants {
		tmc, err := p.tmcDB.GetByEngine(tenant.Engine)
		if err != nil {
			return "", err
		}
		if tmc == nil {
			continue
		}
		if strings.EqualFold(tmc.ServiceType, string(instance.ServiceTypeMicroService)) {
			return p.getProjectIDByTenant(tenant)
		}
	}
	return "", fmt.Errorf("tenant not found from group %q", group)
}

func (p *provider) getProjectIDByTenant(tenant *instancedb.InstanceTenant) (string, error) {
	if len(tenant.Options) <= 0 {
		return "", fmt.Errorf("fail to find project id by tenant %q", tenant.ID)
	}
	options := make(map[string]interface{})
	json.Unmarshal([]byte(tenant.Options), &options)
	pid := options["projectId"]
	if pid == nil {
		return "", fmt.Errorf("fail to find project id by tenant %q", tenant.ID)
	}
	return fmt.Sprint(pid), nil
}
