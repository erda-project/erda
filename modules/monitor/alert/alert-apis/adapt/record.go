package adapt

import (
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"net/url"
)

type (
	// AlertRecord .
	AlertRecord struct {
		GroupID       string `json:"groupId,omitempty"`
		Scope         string `json:"scope,omitempty"`
		ScopeKey      string `json:"scopeKey,omitempty"`
		AlertGroup    string `json:"alertGroup,omitempty"`
		Title         string `json:"title,omitempty"`
		AlertState    string `json:"alertState,omitempty"`
		AlertType     string `json:"alertType,omitempty"`
		AlertIndex    string `json:"alertIndex,omitempty"`
		ExpressionKey string `json:"expressionKey,omitempty"`
		AlertID       uint64 `json:"alertId,omitempty"`
		AlertName     string `json:"alertName,omitempty"`
		RuleID        uint64 `json:"ruleId,omitempty"`
		ProjectID     uint64 `json:"projectId,omitempty"`
		IssueID       uint64 `json:"issueId,omitempty"`
		HandleState   string `json:"handleState,omitempty"`
		HandlerID     string `json:"handlerId,omitempty"`
		AlertTime     int64  `json:"alertTime,omitempty"`
		HandleTime    int64  `json:"handleTime,omitempty"`
		CreateTime    int64  `json:"createTime,omitempty"`
		UpdateTime    int64  `json:"updateTime,omitempty"`
	}

	// AlertRecordAttr .
	AlertRecordAttr struct {
		AlertState  []*DisplayKey `json:"alertState"`
		AlertType   []*DisplayKey `json:"alertType"`
		HandleState []*DisplayKey `json:"handleState"`
	}

	// AlertHistory .
	AlertHistory struct {
		GroupID    string `json:"groupId"`
		Timestamp  int64  `json:"timestamp"`
		AlertState string `json:"alertState"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		DisplayURL string `json:"displayUrl"`
	}

	AlertState string
)

const (
	AlertStateAlert   AlertState = "alert"   // 告警
	AlertStateRecover AlertState = "recover" // 恢复
)

var (
	alertStates = []AlertState{AlertStateAlert, AlertStateRecover}
)

// GetOrgAlertRecordAttr 获取企业告警记录的属性
func (a *Adapt) GetOrgAlertRecordAttr(code i18n.LanguageCodes) (*AlertRecordAttr, error) {
	result, err := a.GetAlertRecordAttr(code, orgScope)
	if err != nil {
		return nil, err
	}
	result.AlertType = append(result.AlertType, &DisplayKey{orgCustomizeType, a.t.Text(code, orgCustomizeType)})
	return result, nil
}

// GetAlertRecordAttr 获取告警记录的属性
func (a *Adapt) GetAlertRecordAttr(code i18n.LanguageCodes, scope string) (*AlertRecordAttr, error) {
	// 查询告警类型
	alertTypes, err := a.db.AlertRule.DistinctAlertTypeByScope(scope)
	if err != nil {
		return nil, err
	}
	// 包装属性
	attr := new(AlertRecordAttr)
	for _, state := range alertStates {
		attr.AlertState = append(attr.AlertState, &DisplayKey{string(state), a.t.Text(code, string(state))})
	}
	for _, state := range apistructs.IssueBugStates {
		attr.HandleState = append(attr.HandleState, &DisplayKey{string(state), a.t.Text(code, string(state))})
	}
	for _, typ := range alertTypes {
		attr.AlertType = append(attr.AlertType, &DisplayKey{typ, a.t.Text(code, typ)})
	}
	return attr, nil
}

// QueryOrgAlertRecord 查询企业告警记录
func (a *Adapt) QueryOrgAlertRecord(lang i18n.LanguageCodes, orgID string,
	alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string, pageNo, pageSize uint) (
	[]*AlertRecord, error) {
	return a.QueryAlertRecord(
		lang, orgScope, orgID, alertGroups, alertStates, alertTypes, handleStates, handlerIDs, pageNo, pageSize)
}

// QueryAlertRecord 查询告警记录
func (a *Adapt) QueryAlertRecord(lang i18n.LanguageCodes, scope, scopeKey string,
	alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string, pageNo, pageSize uint) (
	[]*AlertRecord, error) {
	// 获取记录列表
	records, err := a.db.AlertRecord.QueryByCondition(
		scope, scopeKey, alertGroups, alertStates, alertTypes, handleStates, handlerIDs, pageNo, pageSize)
	if err != nil {
		return nil, err
	}
	// 翻译
	result := make([]*AlertRecord, 0)
	for _, record := range records {
		item := (&AlertRecord{}).FromModel(record)
		item.GroupID = url.QueryEscape(item.GroupID)
		item.AlertType = a.t.Text(lang, record.AlertType)
		result = append(result, item)
	}
	return result, nil
}

// CountOrgAlertRecord 统计企业告警记录
func (a *Adapt) CountOrgAlertRecord(
	orgID string, alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string) (int, error) {
	return a.CountAlertRecord(orgScope, orgID, alertGroups, alertStates, alertTypes, handleStates, handlerIDs)
}

// CountAlertRecord 统计告警记录
func (a *Adapt) CountAlertRecord(
	scope, scopeKey string, alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string) (int, error) {
	return a.db.AlertRecord.CountByCondition(
		scope, scopeKey, alertGroups, alertStates, alertTypes, handleStates, handlerIDs)
}

// GetOrgAlertRecord 获取企业告警记录
func (a *Adapt) GetOrgAlertRecord(lang i18n.LanguageCodes, orgID, groupID string) (*AlertRecord, error) {
	record, err := a.GetAlertRecord(lang, groupID)
	if err != nil {
		return nil, err
	} else if record == nil {
		return nil, nil
	}
	if record.Scope != orgScope || record.ScopeKey != orgID {
		return nil, nil
	}
	return record, nil
}

// GetAlertRecord 获取告警记录
func (a *Adapt) GetAlertRecord(lang i18n.LanguageCodes, groupID string) (*AlertRecord, error) {
	groupID, err := url.QueryUnescape(groupID)
	if err != nil {
		return nil, err
	}

	record, err := a.db.AlertRecord.GetByGroupID(groupID)
	if err != nil {
		return nil, err
	} else if record == nil {
		return nil, nil
	}
	result := (&AlertRecord{}).FromModel(record)
	result.GroupID = url.QueryEscape(result.GroupID)
	result.AlertType = a.t.Text(lang, record.AlertType)
	return result, nil
}

// QueryOrgAlertHistory 查询企业告警记录
func (a *Adapt) QueryOrgAlertHistory(
	lang i18n.LanguageCodes, orgID, groupID string, start, end int64, limit uint) ([]*AlertHistory, error) {
	// 数据校验
	record, err := a.GetOrgAlertRecord(lang, orgID, groupID)
	if err != nil {
		return nil, err
	} else if record == nil {
		return nil, nil
	}
	// 获取告警历史
	return a.QueryAlertHistory(lang, groupID, start, end, limit)
}

// QueryAlertHistory 查询告警记录
func (a *Adapt) QueryAlertHistory(lang i18n.LanguageCodes, groupID string, start, end int64, limit uint) ([]*AlertHistory, error) {
	groupID, err := url.QueryUnescape(groupID)
	if err != nil {
		return nil, err
	}

	// 查询告警记录
	histories, err := a.cql.AlertHistory.QueryAlertHistory(groupID, start, end, limit)
	if err != nil {
		return nil, err
	}
	// 翻译
	result := make([]*AlertHistory, 0)
	for _, history := range histories {
		item := (&AlertHistory{}).FromModel(history)
		result = append(result, item)
	}
	return result, nil
}

// CreateOrgAlertIssue 创建企业告警工单
func (a *Adapt) CreateOrgAlertIssue(orgID, userID, groupID string, issue apistructs.IssueCreateRequest) (uint64, error) {
	// 数据校验
	record, err := a.GetOrgAlertRecord(i18n.LanguageCodes{}, orgID, groupID)
	if err != nil {
		return 0, err
	} else if record == nil || record.IssueID != 0 {
		return 0, nil
	}
	// 创建issue
	issue.Creator = userID
	return a.CreateAlertIssue(groupID, issue)
}

// CreateAlertIssue 创建告警工单
func (a *Adapt) CreateAlertIssue(groupID string, issue apistructs.IssueCreateRequest) (uint64, error) {
	groupID, err := url.QueryUnescape(groupID)
	if err != nil {
		return 0, err
	}

	issue.Source = "alert"
	issue.IterationID = -1
	issue.Type = apistructs.IssueTypeTicket
	// 创建issue
	issueID, err := a.bdl.CreateIssueTicket(issue)
	if err != nil {
		return 0, err
	}
	// 修改记录
	return issueID, a.db.AlertRecord.UpdateHandle(groupID, issueID, issue.Assignee, string(apistructs.IssueStateOpen))
}

// UpdateOrgAlertIssue 修改企业告警工单
func (a *Adapt) UpdateOrgAlertIssue(orgID, groupID string, issue apistructs.IssueUpdateRequest) error {
	if issue.IterationID == nil {
		issue.IterationID = new(int64)
	}
	*issue.IterationID = -1
	// 数据校验
	record, err := a.GetOrgAlertRecord(i18n.LanguageCodes{}, orgID, groupID)
	if err != nil {
		return err
	} else if record == nil {
		return nil
	}
	// 创建issue
	return a.UpdateAlertIssue(groupID, record.IssueID, issue)
}

// UpdateAlertIssue 修改告警工单
func (a *Adapt) UpdateAlertIssue(groupID string, issueID uint64, issue apistructs.IssueUpdateRequest) error {
	groupID, err := url.QueryUnescape(groupID)
	if err != nil {
		return err
	}

	// 创建issue
	err = a.bdl.UpdateIssueTicket(issue, issueID)
	if err != nil {
		return err
	}
	// 修改记录
	return a.db.AlertRecord.UpdateHandle(groupID, issueID, *issue.Assignee, string(*issue.State))
}
