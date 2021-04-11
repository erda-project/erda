package apis

import (
	"net/http"
	"time"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	ext "github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/utils"
)

type clusterReq struct {
	ClusterName string   `json:"clusterName"`
	HostIPs     []string `json:"hostIPs"`
}

type queryOrgHostsAlertRecordReq struct {
	Clusters []*clusterReq `json:"clusters" query:"clusters"`
	QueryOrgAlertRecordReq
}

type QueryOrgAlertRecordReq struct {
	AlertGroup  []string `query:"alertGroup"`
	AlertState  []string `query:"alertState"`
	AlertType   []string `query:"alertType"`
	HandleState []string `query:"handleState"`
	HandlerID   []string `query:"handlerId"`
	PageNo      uint     `query:"pageNo" validate:"gte=0"`
	PageSize    uint     `query:"pageSize" validate:"gte=0,lte=100"`
}

// Query enterprise alarm record attributes (used for translation of front-end query conditions)
func (p *provider) getOrgAlertRecordAttr(r *http.Request) interface{} {
	data, err := p.a.GetOrgAlertRecordAttr(api.Language(r))
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) queryOrgHostsAlertRecord(r *http.Request, params queryOrgHostsAlertRecordReq) interface{} {
	params.AlertGroup = make([]string, 0)
	for _, cluster := range params.Clusters {
		for _, hostIP := range cluster.HostIPs {
			params.AlertGroup = append(params.AlertGroup, cluster.ClusterName+"-"+hostIP)
		}
	}
	return p.queryOrgAlertRecord(r, params.QueryOrgAlertRecordReq)
}

func (p *provider) queryOrgAlertRecord(r *http.Request, params QueryOrgAlertRecordReq) interface{} {
	if params.PageNo == 0 {
		params.PageNo = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}
	list, err := p.a.QueryOrgAlertRecord(api.Language(r), api.OrgID(r),
		params.AlertGroup, params.AlertState, params.AlertType, params.HandleState, params.HandlerID,
		params.PageNo, params.PageSize)
	if err != nil {
		return api.Errors.Internal(err)
	}
	count, err := p.a.CountOrgAlertRecord(
		api.OrgID(r), params.AlertGroup, params.AlertState, params.AlertType, params.HandleState, params.HandlerID)
	if err != nil {
		return api.Errors.Internal(err)
	}

	userIDMap := make(map[string]bool)
	for _, item := range list {
		if item.HandlerID == "" {
			continue
		}
		userIDMap[item.HandlerID] = true
	}
	userIDs := make([]string, 0)
	for key := range userIDMap {
		userIDs = append(userIDs, key)
	}
	return ext.SuccessExt(&listResult{list, count}, userIDs)
}

func (p *provider) getOrgAlertRecord(r *http.Request, params struct {
	GroupID string `query:"groupId" validate:"required"`
}) interface{} {
	data, err := p.a.GetOrgAlertRecord(api.Language(r), api.OrgID(r), params.GroupID)
	if err != nil {
		return api.Errors.Internal(err)
	} else if data == nil {
		return api.Success(nil)
	}

	// Obtain the corresponding projectId according to the issue
	if data.IssueID != 0 {
		issue, err := p.bdl.GetIssue(data.IssueID)
		if err != nil {
			return api.Errors.Internal(err)
		} else if issue != nil {
			data.ProjectID = issue.ProjectID
		}
	}
	return api.Success(data)
}

func (p *provider) queryOrgAlertHistory(r *http.Request, params struct {
	GroupID string `query:"groupId" validate:"required"`
	Start   int64  `query:"start" validate:"gte=0"`
	End     int64  `query:"end" validate:"gte=0"`
	Limit   uint   `query:"limit" validate:"gte=0"`
}) interface{} {
	if params.End == 0 {
		params.End = utils.ConvertTimeToMS(time.Now())
	}
	if params.Start == 0 {
		params.Start = params.End - time.Hour.Milliseconds()
	}
	if params.End < params.Start {
		return api.Success([]*adapt.AlertHistory{})
	}
	if params.Limit == 0 {
		params.Limit = 50
	}
	data, err := p.a.QueryOrgAlertHistory(
		api.Language(r), api.OrgID(r), params.GroupID, params.Start, params.End, params.Limit)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) createOrgAlertIssue(r *http.Request, params struct {
	apistructs.IssueCreateRequest
	GroupID string `query:"groupId" validate:"required"`
}) interface{} {
	issueId, err := p.a.CreateOrgAlertIssue(api.OrgID(r), api.UserID(r), params.GroupID, params.IssueCreateRequest)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(issueId)
}

func (p *provider) updateOrgAlertIssue(r *http.Request, params struct {
	apistructs.IssueUpdateRequest
	GroupID string `query:"groupId" validate:"required"`
}) interface{} {
	if err := p.a.UpdateOrgAlertIssue(api.OrgID(r), params.GroupID, params.IssueUpdateRequest); err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}
