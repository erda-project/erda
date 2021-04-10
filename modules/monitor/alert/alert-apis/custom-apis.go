package apis

import (
	"fmt"
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"net/http"
)

func (p *provider) queryCustomizeMetric(r *http.Request, params struct {
	Names   []string `query:"name"`
	Scope   string   `query:"scope" validate:"required"`
	ScopeID string   `query:"scopeId" validate:"required"`
}) interface{} {
	data, err := p.a.CustomizeMetrics(api.Language(r), params.Scope, params.ScopeID, params.Names)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) queryCustomizeNotifyTarget(r *http.Request) interface{} {
	return api.Success(map[string]interface{}{
		"targets": p.a.NotifyTargetsKeys(api.Language(r), api.OrgID(r)),
	})
}

func (p *provider) queryCustomizeAlert(r *http.Request, params struct {
	Scope    string `query:"scope" validate:"required"`
	ScopeID  string `query:"scopeId" validate:"required"`
	PageNo   int    `query:"pageNo" validate:"gte=1"`
	PageSize int    `query:"pageSize" validate:"gte=1,lte=100"`
}) interface{} {
	alert, total, err := p.a.CustomizeAlerts(api.Language(r), params.Scope, params.ScopeID, params.PageNo, params.PageSize)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"total": total,
		"list":  alert,
	})
}

func (p *provider) getCustomizeAlert(params struct {
	ID uint64 `param:"id" validate:"required,gt=0"`
}) interface{} {
	alert, err := p.a.CustomizeAlert(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(alert)
}

func (p *provider) getCustomizeAlertDetail(params struct {
	ID uint64 `param:"id" validate:"required,gt=0"`
}) interface{} {
	alert, err := p.a.CustomizeAlertDetail(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(alert)
}

func (p *provider) createCustomizeAlert(alert adapt.CustomizeAlertDetail, r *http.Request) interface{} {
	alert.Lang = api.Language(r)
	err := p.checkCustomizeAlert(&alert)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	id, err := p.a.CreateCustomizeAlert(&alert)
	if err != nil {
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists("alert")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(id)
}

func (p *provider) queryOrgDashboardByAlert(r *http.Request, alert *adapt.CustomizeAlertDetail) interface{} {
	orgID := api.OrgID(r)
	if alert.AlertType == "" {
		alert.AlertType = "org_customize"
	}
	_, err := p.bdl.GetOrg(orgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	alert.Lang = api.Language(r)
	alert.AlertScope = "org"
	alert.AlertScopeID = orgID
	if alert.AlertScope == "" {
		return fmt.Errorf("alert scope must not be empty")
	}
	if alert.AlertScopeID == "" {
		return fmt.Errorf("alert scope id must not be empty")
	}

	dash, err := adapt.NewDashboard(p.a).GenerateDashboardPreView(alert)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(dash)
}

func (p *provider) queryDashboardByAlert(r *http.Request, alert *adapt.CustomizeAlertDetail) interface{} {
	alert.Lang = api.Language(r)
	if alert.AlertScope == "" {
		return fmt.Errorf("alert scope must not be empty")
	}
	if alert.AlertScopeID == "" {
		return fmt.Errorf("alert scope id must not be empty")
	}

	dash, err := adapt.NewDashboard(p.a).GenerateDashboardPreView(alert)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(dash)
}

func (p *provider) updateCustomizeAlert(params struct {
	ID uint64 `param:"id" validate:"required,gt=0"`
}, alert adapt.CustomizeAlertDetail) interface{} {
	err := p.checkCustomizeAlert(&alert)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	alert.ID = params.ID
	err = p.a.UpdateCustomizeAlert(&alert)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) checkCustomizeAlert(alert *adapt.CustomizeAlertDetail) error {
	if alert.Name == "" {
		return fmt.Errorf("alert name must not be empty")
	}
	if alert.AlertScope == "" {
		return fmt.Errorf("alert scope must not be empty")
	}
	if alert.AlertScopeID == "" {
		return fmt.Errorf("alert scope id must not be empty")
	}
	if len(alert.Rules) == 0 {
		return fmt.Errorf("alert rules id must not be empty")
	}
	if len(alert.Notifies) == 0 {
		return fmt.Errorf("alert notifies must not be empty")
	}
	// 必须包含ticket类型的通知方式，用于告警历史展示
	hasTicket := false
	for _, notify := range alert.Notifies {
		for _, target := range notify.Targets {
			if target == "ticket" {
				hasTicket = true
				break
			}
		}
	}
	if !hasTicket {
		return fmt.Errorf("alert notifies must has ticket")
	}
	return nil
}

func (p *provider) updateCustomizeAlertEnable(params struct {
	ID     uint64 `param:"id" validate:"required,gt=0"`
	Enable bool   `param:"enable"`
}) interface{} {
	err := p.a.UpdateCustomizeAlertEnable(params.ID, params.Enable)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) deleteCustomizeAlert(params struct {
	ID uint64 `param:"id" validate:"required,gt=0"`
}) interface{} {
	data, _ := p.a.CustomizeAlert(params.ID)
	err := p.a.DeleteCustomizeAlert(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data != nil {
		return api.Success(map[string]interface{}{
			"name": data.Name,
		})
	}
	return api.Success(nil)
}
