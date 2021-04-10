package apis

import (
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"net/http"
	"strconv"
)

func (p *provider) queryOrgAlertRule(r *http.Request) interface{} {
	orgID := api.OrgID(r)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	data, err := p.a.QueryOrgAlertRule(api.Language(r), id)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) queryOrgAlert(r *http.Request, params struct {
	PageNo   int `query:"pageNo" validate:"gte=1"`
	PageSize int `query:"pageSize" validate:"gte=1,lte=100"`
}) interface{} {
	orgID := api.OrgID(r)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	data, err := p.a.QueryOrgAlert(api.Language(r), id, uint64(params.PageNo), uint64(params.PageSize))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data == nil {
		data = make([]*adapt.Alert, 0)
	}
	total, err := p.a.CountOrgAlert(id)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"list":  data,
		"total": total,
	})
}

func (p *provider) getOrgAlertDetail(r *http.Request, params struct {
	ID int `param:"id" validate:"gte=1"`
}) interface{} {
	orgID := api.OrgID(r)
	data, err := p.a.GetOrgAlertDetail(api.Language(r), uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data.AlertScope != "org" || data.AlertScopeID != orgID {
		return api.Success(nil)
	}
	return api.Success(data)
}

func (p *provider) createOrgAlert(r *http.Request, alert adapt.Alert) interface{} {
	orgID := api.OrgID(r)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	if len(alert.ClusterNames) <= 0 {
		return api.Errors.MissingParameter("cluster names")
	}
	if !p.checkOrgClusterNames(id, alert.ClusterNames) {
		return api.Errors.AccessDenied()
	}
	aid, err := p.a.CreateOrgAlert(&alert, orgID)
	if err != nil {
		if adapt.IsInvalidParameterError(err) {
			return api.Errors.InvalidParameter(err)
		}
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists(err)
		}
		return api.Errors.Internal(err)
	}
	return api.Success(aid)
}

// checkOrgClusterNames .
func (p *provider) checkOrgClusterNames(orgID uint64, clusters []string) bool {
	rels, err := p.cmdb.QueryAllOrgClusterRelation()
	if err != nil {
		p.L.Errorf("fail to QueryAllOrgClusterRelation(): %s", err)
		return false
	}
	clustersMap := make(map[string]bool)
	for _, item := range rels {
		if item.OrgID == orgID {
			clustersMap[item.ClusterName] = true
		}
	}
	for _, clusterName := range clusters {
		if _, ok := clustersMap[clusterName]; !ok {
			return false
		}
	}
	return true
}

// createOrgAlert .
func (p *provider) updateOrgAlert(r *http.Request, params struct {
	ID int `param:"id" validate:"required,gt=0"`
}, alert adapt.Alert) interface{} {
	orgID := api.OrgID(r)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	if len(alert.ClusterNames) <= 0 {
		return api.Errors.MissingParameter("cluster names")
	}
	if !p.checkOrgClusterNames(id, alert.ClusterNames) {
		return api.Errors.AccessDenied()
	}
	if err := p.a.UpdateOrgAlert(uint64(params.ID), &alert, orgID); err != nil {
		if adapt.IsInvalidParameterError(err) {
			return api.Errors.InvalidParameter(err)
		}
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists(err)
		}
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) updateOrgAlertEnable(r *http.Request, params struct {
	ID     int  `param:"id" validate:"required,gt=0"`
	Enable bool `query:"enable"`
}) interface{} {
	orgID := api.OrgID(r)
	if len(orgID) <= 0 {
		return api.Errors.InvalidParameter("Org-ID not exist")
	}
	err := p.a.UpdateOrgAlertEnable(uint64(params.ID), params.Enable, orgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) deleteOrgAlert(r *http.Request, params struct {
	ID int `param:"id" validate:"required,gt=0"`
}) interface{} {
	orgID := api.OrgID(r)
	if len(orgID) <= 0 {
		return api.Errors.InvalidParameter("Org-ID not exist")
	}
	data, _ := p.a.GetAlert(api.Language(r), uint64(params.ID))
	err := p.a.DeleteOrgAlert(uint64(params.ID), orgID)
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
