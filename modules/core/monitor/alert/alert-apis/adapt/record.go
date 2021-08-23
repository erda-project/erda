// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adapt

import (
	"fmt"
	"net/url"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda/apistructs"
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

// GetOrgAlertRecordAttr Get the attributes of enterprise alarm records
func (a *Adapt) GetOrgAlertRecordAttr(code i18n.LanguageCodes) (*pb.AlertRecordAttr, error) {
	result, err := a.GetAlertRecordAttr(code, orgScope)
	if err != nil {
		return nil, err
	}
	result.AlertType = append(result.AlertType, &pb.DisplayKey{Key: orgCustomizeType, Display: a.t.Text(code, orgCustomizeType)})
	return result, nil
}

// GetAlertRecordAttr Get the attributes of the alarm record
func (a *Adapt) GetAlertRecordAttr(code i18n.LanguageCodes, scope string) (*pb.AlertRecordAttr, error) {
	// Query alarm type
	alertTypes, err := a.db.AlertRule.DistinctAlertTypeByScope(scope)
	if err != nil {
		return nil, err
	}
	attr := new(pb.AlertRecordAttr)
	for _, state := range alertStates {
		attr.AlertState = append(attr.AlertState, &pb.DisplayKey{Key: string(state), Display: a.t.Text(code, string(state))})
	}
	for _, state := range apistructs.IssueBugStates {
		attr.HandleState = append(attr.HandleState, &pb.DisplayKey{Key: string(state), Display: a.t.Text(code, string(state))})
	}
	for _, typ := range alertTypes {
		attr.AlertType = append(attr.AlertType, &pb.DisplayKey{Key: typ, Display: a.t.Text(code, typ)})
	}
	return attr, nil
}

// QueryOrgAlertRecord
func (a *Adapt) QueryOrgAlertRecord(lang i18n.LanguageCodes, orgID string,
	alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string, pageNo, pageSize uint) (
	[]*pb.AlertRecord, error) {
	return a.QueryAlertRecord(
		lang, orgScope, orgID, alertGroups, alertStates, alertTypes, handleStates, handlerIDs, pageNo, pageSize)
}

// QueryAlertRecord
func (a *Adapt) QueryAlertRecord(lang i18n.LanguageCodes, scope, scopeKey string,
	alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string, pageNo, pageSize uint) (
	[]*pb.AlertRecord, error) {
	// get list
	records, err := a.db.AlertRecord.QueryByCondition(
		scope, scopeKey, alertGroups, alertStates, alertTypes, handleStates, handlerIDs, pageNo, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.AlertRecord, 0)
	for _, record := range records {
		item := ToPBAlertRecord(record)
		item.GroupId = url.QueryEscape(item.GroupId)
		item.AlertType = a.t.Text(lang, record.AlertType)
		result = append(result, item)
	}
	return result, nil
}

// CountOrgAlertRecord
func (a *Adapt) CountOrgAlertRecord(
	orgID string, alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string) (int, error) {
	return a.CountAlertRecord(orgScope, orgID, alertGroups, alertStates, alertTypes, handleStates, handlerIDs)
}

// CountAlertRecord
func (a *Adapt) CountAlertRecord(
	scope, scopeKey string, alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string) (int, error) {
	return a.db.AlertRecord.CountByCondition(
		scope, scopeKey, alertGroups, alertStates, alertTypes, handleStates, handlerIDs)
}

// GetOrgAlertRecord
func (a *Adapt) GetOrgAlertRecord(lang i18n.LanguageCodes, orgID, groupID string) (*pb.AlertRecord, error) {
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

// GetAlertRecord
func (a *Adapt) GetAlertRecord(lang i18n.LanguageCodes, groupID string) (*pb.AlertRecord, error) {
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
	result := ToPBAlertRecord(record)
	result.GroupId = url.QueryEscape(result.GroupId)
	result.AlertType = a.t.Text(lang, record.AlertType)
	return result, nil
}

// QueryOrgAlertHistory
func (a *Adapt) QueryOrgAlertHistory(
	lang i18n.LanguageCodes, orgID, groupID string, start, end int64, limit uint) ([]*pb.AlertHistory, error) {
	record, err := a.GetOrgAlertRecord(lang, orgID, groupID)
	if err != nil {
		return nil, err
	} else if record == nil {
		return nil, nil
	}
	return a.QueryAlertHistory(lang, groupID, start, end, limit)
}

// QueryAlertHistory
func (a *Adapt) QueryAlertHistory(lang i18n.LanguageCodes, groupID string, start, end int64, limit uint) ([]*pb.AlertHistory, error) {
	groupID, err := url.QueryUnescape(groupID)
	if err != nil {
		return nil, err
	}

	histories, err := a.cql.AlertHistory.QueryAlertHistory(groupID, start, end, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*pb.AlertHistory, 0)
	for _, history := range histories {
		item := ToDBAlertHistory(history)
		result = append(result, item)
	}
	return result, nil
}

// CreateOrgAlertIssue
func (a *Adapt) CreateOrgAlertIssue(orgID, userID, groupID string, issue apistructs.IssueCreateRequest) (uint64, error) {
	record, err := a.GetOrgAlertRecord(i18n.LanguageCodes{}, orgID, groupID)
	if err != nil {
		return 0, err
	} else if record == nil || record.IssueId != 0 {
		return 0, nil
	}
	issue.Creator = userID
	return a.CreateAlertIssue(groupID, issue)
}

// CreateAlertIssue
func (a *Adapt) CreateAlertIssue(groupID string, issue apistructs.IssueCreateRequest) (uint64, error) {
	groupID, err := url.QueryUnescape(groupID)
	if err != nil {
		return 0, err
	}

	issue.Source = "alert"
	issue.IterationID = -1
	issue.Type = apistructs.IssueTypeTicket
	issueID, err := a.bdl.CreateIssueTicket(issue)
	if err != nil {
		return 0, err
	}
	return issueID, a.db.AlertRecord.UpdateHandle(groupID, issueID, issue.Assignee, string(apistructs.IssueStateOpen))
}
func (a *Adapt) UpdateOrgAlertIssue(orgID, groupID string, issue apistructs.IssueUpdateRequest) error {
	if issue.IterationID == nil {
		issue.IterationID = new(int64)
	}
	*issue.IterationID = -1
	record, err := a.GetOrgAlertRecord(i18n.LanguageCodes{}, orgID, groupID)
	if err != nil {
		return err
	} else if record == nil {
		return nil
	}
	return a.UpdateAlertIssue(groupID, record.IssueId, issue)
}

// UpdateAlertIssue
func (a *Adapt) UpdateAlertIssue(groupID string, issueID uint64, issue apistructs.IssueUpdateRequest) error {
	groupID, err := url.QueryUnescape(groupID)
	if err != nil {
		return err
	}

	err = a.bdl.UpdateIssueTicket(issue, issueID)
	if err != nil {
		return err
	}
	return a.db.AlertRecord.UpdateHandle(groupID, issueID, *issue.Assignee, fmt.Sprintf("%v", *issue.State))
}
