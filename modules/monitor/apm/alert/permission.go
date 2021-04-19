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
	"github.com/erda-project/erda-infra/providers/httpserver"
	apm "github.com/erda-project/erda/modules/monitor/apm/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) getPermissionByTenantGroup(action permission.Action) httpserver.Interceptor {
	return permission.Intercepter(
		permission.ScopeProject, permission.TenantGroupFromParams(p.authDb),
		apm.MonitorProjectAlert, action,
	)
}
