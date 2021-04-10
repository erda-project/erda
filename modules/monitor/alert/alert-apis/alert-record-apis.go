package apis

import (
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"github.com/erda-project/erda/modules/monitor/utils"
	"net/http"
	"time"
)

// 查询告警记录属性（用于前端查询条件翻译）
func (p *provider) getAlertRecordAttr(r *http.Request, params struct {
	Scope string `query:"scope" validate:"required"`
}) interface{} {
	data, err := p.a.GetAlertRecordAttr(api.Language(r), params.Scope)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

// 查询告警记录
func (p *provider) queryAlertRecord(r *http.Request, params struct {
	Scope       string   `query:"scope" validate:"required"`
	ScopeKey    string   `query:"scopeKey" validate:"required"`
	AlertGroup  []string `query:"alertGroup"`
	AlertState  []string `query:"alertState"`
	AlertType   []string `query:"alertType"`
	HandleState []string `query:"handleState"`
	HandlerID   []string `query:"handlerId"`
	PageNo      uint     `query:"pageNo" validate:"gte=0"`
	PageSize    uint     `query:"pageSize" validate:"gte=0,lte=100"`
}) interface{} {
	if params.PageNo == 0 {
		params.PageNo = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}
	list, err := p.a.QueryAlertRecord(api.Language(r), params.Scope, params.ScopeKey,
		params.AlertGroup, params.AlertState, params.AlertType, params.HandleState, params.HandlerID,
		params.PageNo, params.PageSize)
	if err != nil {
		return api.Errors.Internal(err)
	}
	count, err := p.a.CountAlertRecord(params.Scope, params.ScopeKey,
		params.AlertGroup, params.AlertState, params.AlertType, params.HandleState, params.HandlerID)
	if err != nil {
		return api.Errors.Internal(err)
	}

	return api.Success(&listResult{list, count})
}

// 获取告警记录
func (p *provider) getAlertRecord(r *http.Request, params struct {
	GroupID string `query:"groupId" validate:"required"`
}) interface{} {
	data, err := p.a.GetAlertRecord(api.Language(r), params.GroupID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

// 查询告警历史
func (p *provider) queryAlertHistory(r *http.Request, params struct {
	GroupID string `query:"groupId" validate:"required"`
	Start   int64  `query:"start" validate:"gte=0"`
	End     int64  `query:"end" validate:"gte=0"`
	Limit   uint   `query:"limit" validate:"gte=0"`
}) interface{} {
	if params.End < params.Start {
		return api.Success([]*adapt.AlertHistory{})
	}
	if params.End == 0 {
		params.End = utils.ConvertTimeToMS(time.Now())
	}
	if params.Start == 0 {
		params.Start = params.End - time.Hour.Milliseconds()
	}
	if params.Limit == 0 {
		params.Limit = 50
	}
	data, err := p.a.QueryAlertHistory(api.Language(r), params.GroupID, params.Start, params.End, params.Limit)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

// 创建企业告警工单
func (p *provider) createAlertIssue(r *http.Request, params struct {
	apistructs.IssueCreateRequest
	GroupID string `query:"groupId" validate:"required"`
}) interface{} {
	if _, err := p.a.CreateAlertIssue(params.GroupID, params.IssueCreateRequest); err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

// 修改企业告警工单
func (p *provider) updateAlertIssue(r *http.Request, params struct {
	apistructs.IssueUpdateRequest
	GroupID string `query:"groupId" validate:"required"`
	IssueID uint64 `query:"issueId" validate:"required"`
}) interface{} {
	if err := p.a.UpdateAlertIssue(params.GroupID, params.IssueID, params.IssueUpdateRequest); err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}
