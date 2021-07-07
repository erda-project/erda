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
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/jinzhu/gorm"
	"strings"

	monitor "github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	alert "github.com/erda-project/erda-proto-go/msp/apm/alert/pb"
	mperm "github.com/erda-project/erda/modules/msp/instance/permission"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type provider struct {
	C            *config
	DB           *gorm.DB
	Register     transport.Register `autowired:"service-register" optional:"true"`
	Perm         perm.Interface     `autowired:"permission"`
	MPerm        mperm.Interface    `autowired:"msp.permission"`
	alertService *alertService
	Monitor      monitor.AlertServiceServer `autowired:"erda.core.monitor.alert.AlertService" optional:"true"`
	authDb       *db.DB

	microServiceFilterTags map[string]bool
}

type config struct {
	MicroServiceFilterTags string `file:"micro_service_filter_tags"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.microServiceFilterTags = make(map[string]bool)
	for _, k := range strings.Split(p.C.MicroServiceFilterTags, ",") {
		k = strings.TrimSpace(k)
		if len(k) > 0 {
			p.microServiceFilterTags[k] = true
		}
	}
	p.alertService = &alertService{p}
	p.authDb = db.New(ctx.Service("mysql").(mysql.Interface).DB())
	p.alertService = &alertService{
		p: p,
	}
	if p.Register != nil {
		type AlertService = alert.AlertServiceServer
		//alert.RegisterAlertServiceImp(p.Register,p.alertService,apis.Options())
		alert.RegisterAlertServiceImp(p.Register, p.alertService, apis.Options(), p.Perm.Check(
			perm.Method(AlertService.QueryAlertRule, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionList, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.QueryAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionList, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.GetAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionGet, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.CreateAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionCreate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.UpdateAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionUpdate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.UpdateAlertEnable, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionUpdate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.DeleteAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionDelete, perm.FieldValue("tenantGroup")),

			perm.Method(AlertService.QueryCustomizeMetric, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionGet, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.QueryCustomizeNotifyTarget, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionGet, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.QueryCustomizeAlerts, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionList, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.GetCustomizeAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionGet, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.CreateCustomizeAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionCreate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.UpdateCustomizeAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionUpdate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.UpdateCustomizeAlertEnable, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionUpdate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.DeleteCustomizeAlert, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionDelete, perm.FieldValue("tenantGroup")),

			perm.Method(AlertService.GetAlertRecordAttrs, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionList, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.GetAlertRecords, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionList, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.GetAlertRecord, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionGet, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.GetAlertHistories, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionList, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.CreateAlertRecordIssue, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionCreate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.UpdateAlertRecordIssue, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionUpdate, perm.FieldValue("tenantGroup")),
			perm.Method(AlertService.DashboardPreview, perm.ScopeProject, perm.MonitorProjectAlert, perm.ActionUpdate, perm.FieldValue("tenantGroup")),
		))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.alert.AlertService" || ctx.Type() == alert.AlertServiceServerType() || ctx.Type() == alert.AlertServiceHandlerType():
		return p.alertService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.apm.alert", &servicehub.Spec{
		Services:             alert.ServiceNames(),
		Types:                alert.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
