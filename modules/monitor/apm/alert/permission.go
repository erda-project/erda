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
