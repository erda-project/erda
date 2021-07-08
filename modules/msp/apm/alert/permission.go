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

package alert

import (
	"context"
	"fmt"
	table "github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/erda-project/erda/modules/monitor/utils"
)

func (m *alertService) TenantGroupFromParams() func(ctx context.Context, req interface{}) (string, error) {
	return func(ctx context.Context, req interface{}) (string, error) {
		return projectIdFromTenantGroup(ctx, m.p.authDb)
	}
}

func projectIdFromTenantGroup(ctx context.Context, db *table.DB) (string, error) {
	value := QueryValue("tenantGroup")
	tenantGroup, err := value(ctx)
	if err != nil {
		return "", err
	}
	tk, err := db.InstanceTenant.QueryTkByTenantGroup(tenantGroup)
	if err != nil {
		return "", err
	}
	projectId, err := db.Monitor.SelectProjectIdByTk(tk)
	if err != nil {
		return "", err
	}
	return projectId, nil
}

func QueryValue(keys ...string) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		request := utils.GetHttpRequest(ctx)
		params := request.URL.Query()
		for _, key := range keys {
			vals := params[key]
			if len(vals) == 1 {
				val := vals[0]
				if len(val) == 0 {
					continue
				}
				return val, nil
			}
			if len(vals) > 1 {
				return "", fmt.Errorf("too many key %s present", key)
			}
		}
		return "", fmt.Errorf("keys %v not found", keys)
	}
}
