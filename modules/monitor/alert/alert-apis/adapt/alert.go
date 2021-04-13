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

package adapt

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/monitor/utils"
)

type (
	// AlertRule .
	AlertRule struct {
		ID         uint64                 `json:"id"`
		Name       string                 `json:"-"`
		AlertScope string                 `json:"-"`
		AlertType  string                 `json:"-"`
		AlertIndex *DisplayKey            `json:"alertIndex"`
		Template   map[string]interface{} `json:"-"`
		Window     int64                  `json:"window"`
		Functions  []*AlertRuleFunction   `json:"functions"`
		IsRecover  bool                   `json:"isRecover"`
		Attributes map[string]interface{} `json:"-"`
		Version    string                 `json:"-"`
		Enable     bool                   `json:"-"`
		CreateTime int64                  `json:"createTime"`
		UpdateTime int64                  `json:"updateTime"`
	}
	// AlertRuleFunction .
	AlertRuleFunction struct {
		Field      *DisplayKey `json:"field"`
		Aggregator string      `json:"aggregator"`
		Operator   string      `json:"operator"`
		Value      interface{} `json:"value"`
		DataType   string      `json:"dataType"`
		Unit       string      `json:"unit" translate:"true"`
	}
	// AlertTypeRuleResp .
	AlertTypeRuleResp struct {
		AlertTypeRules []*AlertTypeRule `json:"alertTypeRules"`
		Windows        []int64          `json:"windows"`
		Operators      []*Operator      `json:"operators"`
		Aggregator     []*DisplayKey    `json:"aggregator"`
		Silence        []*NotifySilence `json:"silence"`
	}
	// AlertTypeRule .
	AlertTypeRule struct {
		AlertType *DisplayKey  `json:"alertType"`
		Rules     []*AlertRule `json:"rules"`
	}
	// AlertNotify .
	AlertNotify struct {
		ID          uint64                  `json:"id"`
		Type        string                  `json:"type"`
		GroupID     int64                   `json:"groupId"`
		GroupType   string                  `json:"groupType"`
		NotifyGroup *apistructs.NotifyGroup `json:"notifyGroup"`
		DingdingURL string                  `json:"dingdingUrl"`
		Silence     *AlertNotifySilence     `json:"silence"`
		CreateTime  int64                   `json:"createTime"`
		UpdateTime  int64                   `json:"updateTime"`
	}
	// AlertNotifySilence .
	AlertNotifySilence struct {
		Value  int64  `json:"value"`
		Unit   string `json:"unit"`
		Policy string `json:"policy"`
	}
	// AlertExpression .
	AlertExpression struct {
		ID         uint64                     `json:"id"`
		RuleID     uint64                     `json:"ruleId"`
		AlertIndex string                     `json:"alertIndex"`
		Window     int64                      `json:"window"`
		Functions  []*AlertExpressionFunction `json:"functions"`
		IsRecover  bool                       `json:"isRecover"`
		CreateTime int64                      `json:"createTime"`
		UpdateTime int64                      `json:"updateTime"`
	}
	// AlertExpressionFunction .
	AlertExpressionFunction struct {
		Field      string      `json:"field"`
		Aggregator string      `json:"aggregator"`
		Operator   string      `json:"operator"`
		Value      interface{} `json:"value"`
	}
	// Alert .
	Alert struct {
		ID           uint64                 `json:"id"`
		Name         string                 `json:"name"`
		AlertScope   string                 `json:"alertScope"`
		AlertScopeID string                 `json:"alertScopeId"`
		Enable       bool                   `json:"enable"`
		Rules        []*AlertExpression     `json:"rules"`
		Notifies     []*AlertNotify         `json:"notifies"`
		Filters      map[string]interface{} `json:"filters"`
		Attributes   map[string]interface{} `json:"attributes"`
		ClusterNames []string               `json:"clusterNames"`
		Domain       string                 `json:"domain"`
		CreateTime   int64                  `json:"createTime"`
		UpdateTime   int64                  `json:"updateTime"`
	}
)

// NotifyTargetType .
type NotifyTargetType string

// NotifyTargetType values
const (
	SysNotifyTarget                NotifyTargetType = "sys"
	OrgNotifyTarget                NotifyTargetType = "org"
	ProjectNotifyTarget            NotifyTargetType = "project"
	AppNotifyTarget                NotifyTargetType = "app"
	UserNotifyTarget               NotifyTargetType = "user"
	ExternalUserNotifyTarget       NotifyTargetType = "external_user"
	DingdingNotifyTarget           NotifyTargetType = "dingding"
	DingdingWorkNoticeNotifyTarget NotifyTargetType = "dingding_worknotice"
	WebhookNotifyTarget            NotifyTargetType = "webhook"

	dashboardPath = "/dataCenter/customDashboard"
	recordPath    = "/dataCenter/alarm/record"
)

// QueryAlertRule .
func (a *Adapt) QueryAlertRule(lang i18n.LanguageCodes, scope, scopeID string) (*AlertTypeRuleResp, error) {
	rules, err := a.db.AlertRule.QueryEnabledByScope(scope)
	if err != nil {
		return nil, err
	}
	customizeRules, err := a.db.CustomizeAlertRule.QueryEnabledByScope(scope, scopeID)
	if err != nil {
		return nil, err
	}
	rulesMap := make(map[string][]*AlertRule)
	for _, item := range customizeRules {
		rule := (&AlertRule{}).FromCustomizeAlertRule(lang, a.t, item)
		rulesMap[item.AlertType] = append(rulesMap[item.AlertType], rule)
	}
	for _, item := range rules {
		rule := (&AlertRule{}).FromModel(lang, a.t, item)
		rulesMap[item.AlertType] = append(rulesMap[item.AlertType], rule)
	}
	var alertTypeRules []*AlertTypeRule
	for alertType, rules := range rulesMap {
		alertTypeRules = append(alertTypeRules, &AlertTypeRule{
			AlertType: &DisplayKey{
				Key:     alertType,
				Display: a.t.Text(lang, alertType),
			},
			Rules: rules,
		})
	}
	// only show single value operation
	var operators []*Operator
	for _, op := range a.FunctionOperatorKeys(lang) {
		if op.Type == OperatorTypeOne {
			operators = append(operators, op)
		}
	}
	return &AlertTypeRuleResp{
		AlertTypeRules: alertTypeRules,
		Windows:        windowKeys,
		Operators:      operators,
		Aggregator:     a.AggregatorKeys(lang),
		Silence:        a.NotifySilences(lang),
	}, nil
}

// QueryOrgAlertRule .
func (a *Adapt) QueryOrgAlertRule(lang i18n.LanguageCodes, orgID uint64) (*AlertTypeRuleResp, error) {
	return a.QueryAlertRule(lang, "org", strconv.FormatUint(orgID, 10))
}

// QueryAlert .
func (a *Adapt) QueryAlert(code i18n.LanguageCodes, scope, scopeID string, pageNo, pageSize uint64) ([]*Alert, error) {
	alerts, err := a.db.Alert.QueryByScopeAndScopeID(scope, scopeID, pageNo, pageSize)
	if err != nil {
		return nil, err
	}
	var alertIDs []uint64
	for _, alert := range alerts {
		alertIDs = append(alertIDs, alert.ID)
	}
	notifyMap, err := a.getAlertNotifysByAlertIDs(alertIDs)
	if err != nil {
		return nil, err
	}
	var list []*Alert
	for _, item := range alerts {
		alert := (&Alert{}).FromModel(item)
		alert.Notifies = notifyMap[alert.ID]
		list = append(list, alert)
	}
	return list, nil
}

// according to alertID get alert
func (a *Adapt) getAlertNotifysByAlertIDs(alertIDs []uint64) (map[uint64][]*AlertNotify, error) {
	if len(alertIDs) == 0 {
		return nil, nil
	}
	notifies, err := a.db.AlertNotify.QueryByAlertIDs(alertIDs)
	if err != nil {
		return nil, err
	}
	var notifyGroupIDs []string
	for _, notify := range notifies {
		if groupID, ok := utils.GetMapValueUint64(notify.NotifyTarget, "group_id"); ok {
			notifyGroupIDs = append(notifyGroupIDs, strconv.FormatUint(groupID, 10))
		}
	}
	notifyGroupMap := a.getNotifyGroupRelByIDs(notifyGroupIDs)

	notifysMap := make(map[uint64][]*AlertNotify)
	for _, notify := range notifies {
		notifyTarget := (&AlertNotify{}).FromModel(notify, notifyGroupMap)
		if notifyTarget == nil {
			continue
		}
		notifysMap[notify.AlertID] = append(notifysMap[notify.AlertID], notifyTarget)
	}
	return notifysMap, nil
}

// get notify groups
func (a *Adapt) getNotifyGroupRelByIDs(groupIDs []string) map[int64]*apistructs.NotifyGroup {
	if len(groupIDs) == 0 {
		return nil
	}
	notifyGroupsData, err := a.cmdb.QueryNotifyGroup(groupIDs)
	if err != nil {
		a.l.Errorf("fail to query notify group from cmdb error: %s", err)
		return nil
	}
	notifyGroupMap := make(map[int64]*apistructs.NotifyGroup)
	for _, notifyGroup := range notifyGroupsData {
		notifyGroupMap[notifyGroup.ID] = notifyGroup
	}
	return notifyGroupMap
}

// QueryOrgAlert .
func (a *Adapt) QueryOrgAlert(lang i18n.LanguageCodes, orgID uint64, pageNo, pageSize uint64) ([]*Alert, error) {
	scopeID := strconv.FormatUint(orgID, 10)
	alerts, err := a.QueryAlert(lang, "org", scopeID, pageNo, pageSize)
	if err != nil {
		return nil, err
	}
	for _, alert := range alerts {
		if clusterNames, ok := utils.GetMapValueArr(alert.Attributes, "cluster_name"); ok {
			for _, v := range clusterNames {
				if clusterName, ok := v.(string); ok {
					alert.ClusterNames = append(alert.ClusterNames, clusterName)
				}
			}
		} else if clusterName, ok := utils.GetMapValueString(alert.Attributes, "cluster_name"); ok {
			alert.ClusterNames = append(alert.ClusterNames, clusterName)
		}
		alert.Attributes = nil
	}
	return alerts, nil
}

// CountAlert .
func (a *Adapt) CountAlert(scope, scopeID string) (int, error) {
	count, err := a.db.Alert.CountByScopeAndScopeID(scope, scopeID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountOrgAlert .
func (a *Adapt) CountOrgAlert(orgID uint64) (int, error) {
	return a.CountAlert("org", strconv.FormatUint(orgID, 10))
}

// GetAlert .
func (a *Adapt) GetAlert(lang i18n.LanguageCodes, id uint64) (*Alert, error) {
	alert, err := a.db.Alert.GetByID(id)
	if err != nil {
		return nil, err
	} else if alert == nil {
		return nil, nil
	}
	return (&Alert{}).FromModel(alert), nil
}

// GetAlertDetail .
func (a *Adapt) GetAlertDetail(lang i18n.LanguageCodes, id uint64) (*Alert, error) {
	alert, err := a.db.Alert.GetByID(id)
	if err != nil {
		return nil, err
	} else if alert == nil {
		return nil, nil
	}
	// get alert expression
	expressions, err := a.getAlertExpressionsByAlertID(alert.ID)
	if err != nil {
		return nil, err
	}
	// filter alarm expressions that do not match the rule
	var indices []string
	for _, item := range expressions {
		indices = append(indices, item.AlertIndex)
	}
	rulesMap, err := a.getEnabledAlertRulesByScopeAndIndices(lang, alert.AlertScope, alert.AlertScopeID, indices)
	if err != nil {
		return nil, err
	}
	var rules []*AlertExpression
	for _, expression := range expressions {
		if _, ok := rulesMap[expression.AlertIndex]; ok {
			rules = append(rules, expression)
		}
	}
	// get alert notify
	notifys, err := a.getAlertNotifysByAlertID(alert.ID)
	if err != nil {
		return nil, err
	}
	data := (&Alert{}).FromModel(alert)
	data.Rules = rules
	data.Notifies = notifys
	return data, nil
}

func (a *Adapt) getAlertExpressionsByAlertID(alertID uint64) ([]*AlertExpression, error) {
	expressions, err := a.db.AlertExpression.QueryByAlertIDs([]uint64{alertID})
	if err != nil {
		return nil, err
	}
	var list []*AlertExpression
	for _, item := range expressions {
		expression := (&AlertExpression{}).FromModel(item)
		if expression == nil {
			continue
		}
		list = append(list, expression)
	}
	return list, nil
}

func (a *Adapt) getAlertNotifysByAlertID(alertID uint64) ([]*AlertNotify, error) {
	if alertID == 0 {
		return nil, nil
	}
	notifysMap, err := a.getAlertNotifysByAlertIDs([]uint64{alertID})
	if err != nil {
		return nil, err
	}
	return notifysMap[alertID], nil
}

// obtain open alarm rules based on scope and index
func (a *Adapt) getEnabledAlertRulesByScopeAndIndices(lang i18n.LanguageCodes, scope, scopeID string, indices []string) (map[string]*AlertRule, error) {
	if len(indices) == 0 {
		return nil, nil
	}
	rules, err := a.db.AlertRule.QueryEnabledByScopeAndIndices(scope, indices)
	if err != nil {
		return nil, err
	}
	customizeRules, err := a.db.CustomizeAlertRule.QueryEnabledByScopeAndIndices(scope, scopeID, indices)
	if err != nil {
		return nil, err
	}
	rulesMap := make(map[string]*AlertRule)
	for _, item := range rules {
		rulesMap[item.AlertIndex] = (&AlertRule{}).FromModel(lang, a.t, item)
	}
	for _, item := range customizeRules {
		rulesMap[item.AlertIndex] = (&AlertRule{}).FromCustomizeAlertRule(lang, a.t, item)
	}
	return rulesMap, nil
}

// GetOrgAlertDetail .
func (a *Adapt) GetOrgAlertDetail(lang i18n.LanguageCodes, id uint64) (*Alert, error) {
	alert, err := a.GetAlertDetail(lang, id)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, nil
	}

	if clusterNames, ok := utils.GetMapValueArr(alert.Attributes, "cluster_name"); ok {
		for _, v := range clusterNames {
			if clusterName, ok := v.(string); ok {
				alert.ClusterNames = append(alert.ClusterNames, clusterName)
			}
		}
	} else if clusterName, ok := utils.GetMapValueString(alert.Attributes, "cluster_name"); ok {
		alert.ClusterNames = append(alert.ClusterNames, clusterName)
	}
	alert.Attributes = nil
	return alert, nil
}

// CreateAlert .
func (a *Adapt) CreateAlert(alert *Alert) (alertID uint64, err error) {
	tx := a.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			a.l.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	dbAlert, err := tx.Alert.GetByScopeAndScopeIDAndName(alert.AlertScope, alert.AlertScopeID, alert.Name)
	if err != nil {
		return 0, err
	}
	if dbAlert != nil {
		return 0, ErrorAlreadyExists
	}
	alert.Enable = true
	data := alert.ToModel()
	data.ID = 0
	err = tx.Alert.Insert(data)
	if err != nil {
		return 0, nil
	}
	alert.ID = data.ID

	// 创建告警表达式
	var (
		types, indexes []string
		expressionLen  int
	)
	for _, expression := range alert.Rules {
		indexes = append(indexes, expression.AlertIndex)
	}
	ruleMap, err := a.getEnabledAlertRulesByScopeAndIndices(i18n.LanguageCodes{}, alert.AlertScope, alert.AlertScopeID, indexes)
	if err != nil {
		return 0, err
	}
	for _, expression := range alert.Rules {
		rule, ok := ruleMap[expression.AlertIndex]
		if !ok || rule.AlertScope != alert.AlertScope {
			return 0, invalidParameter("rule %s is not scope: %s", rule.AlertIndex.Key, alert.AlertScope)
		}
		exp, err := expression.ToModel(alert, rule)
		if err != nil {
			return 0, err
		}
		exp.ID = 0
		if err := tx.AlertExpression.Insert(exp); err != nil {
			return 0, err
		}
		expressionLen++
		types = append(types, rule.AlertType)
	}
	if expressionLen == 0 {
		return 0, errors.New("expression is not valid")
	}

	// create alert notify
	var (
		silence   *AlertNotifySilence
		notifyLen int
	)
	for _, item := range alert.Notifies {
		silence = item.Silence
		notify := item.ToModel(alert, a.silencePolicies)
		if notify == nil {
			continue
		}
		notify.ID = 0
		if err := tx.AlertNotify.Insert(notify); err != nil {
			return 0, err
		}
		notifyLen++
	}
	if notifyLen == 0 {
		return 0, errors.New("notify is not valid")
	}
	// create ticket alert notify
	notify := a.newTicketAlertNotify(alert.ID, silence)
	if err := tx.AlertNotify.Insert(notify); err != nil {
		return 0, err
	}
	return alert.ID, nil
}

// crate ticket alert notify
func (a *Adapt) newTicketAlertNotify(alertID uint64, silence *AlertNotifySilence) *db.AlertNotify {
	if silence == nil {
		return nil
	}
	t := convertMillisecondByUnit(silence.Value, silence.Unit)
	if t < 0 {
		return nil
	}
	if silence.Policy == "" || !a.silencePolicies[silence.Policy] {
		silence.Policy = fixedSliencePolicy
	}
	return &db.AlertNotify{
		AlertID: alertID,
		NotifyTarget: map[string]interface{}{
			"type": "ticket",
		},
		NotifyTargetID: "",
		Silence:        t,
		SilencePolicy:  silence.Policy,
		Enable:         true,
	}
}

// CreateOrgAlert .
func (a *Adapt) CreateOrgAlert(alert *Alert, orgID string) (alertID uint64, err error) {
	alert.AlertScope = "org"
	alert.AlertScopeID = orgID
	alert.Attributes = make(map[string]interface{})
	alert.Attributes["alert_domain"] = alert.Domain
	alert.Attributes["alert_dashboard_path"] = dashboardPath
	alert.Attributes["alert_record_path"] = recordPath
	alert.Attributes["dice_org_id"] = orgID
	alert.Attributes["cluster_name"] = alert.ClusterNames
	return a.CreateAlert(alert)
}

// UpdateOrgAlert .
func (a *Adapt) UpdateOrgAlert(alertID uint64, alert *Alert, orgID string) error {
	// data authorization
	origin, err := a.db.Alert.GetByID(alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return nil
	}
	if origin.AlertScope != "org" || origin.AlertScopeID != orgID {
		return fmt.Errorf("permission denied")
	}

	// supplement data
	alert.AlertScope = origin.AlertScope
	alert.AlertScopeID = origin.AlertScopeID
	alert.Enable = origin.Enable
	alert.Attributes = make(map[string]interface{})
	for k, v := range origin.Attributes {
		alert.Attributes[k] = v
	}
	if alert.Domain != "" {
		alert.Attributes["alert_domain"] = alert.Domain
	}
	alert.Attributes["alert_dashboard_path"] = dashboardPath
	alert.Attributes["alert_record_path"] = recordPath
	alert.Attributes["cluster_name"] = alert.ClusterNames

	return a.UpdateAlert(alertID, alert)
}

// UpdateAlert .
func (a *Adapt) UpdateAlert(alertID uint64, alert *Alert) (err error) {
	tx := a.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			a.l.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if alert.Name != "" {
		dbAlert, err := tx.Alert.GetByScopeAndScopeIDAndName(alert.AlertScope, alert.AlertScopeID, alert.Name)
		if err != nil {
			return err
		}
		if dbAlert != nil && dbAlert.ID != alertID {
			return ErrorAlreadyExists
		}
	}

	dbAlert, err := tx.Alert.GetByID(alertID)
	if err != nil {
		return err
	}
	if dbAlert == nil {
		return nil
	}
	attributes := make(map[string]interface{})
	for k, v := range dbAlert.Attributes {
		attributes[k] = v
	}
	for k, v := range alert.Attributes {
		attributes[k] = v
	}
	alert.ID = alertID
	alert.Attributes = attributes
	dbAlert = alert.ToModel()
	if err := tx.Alert.Update(dbAlert); err != nil {
		return err
	}

	// modify alert expression
	var (
		indexes       []string
		expressionLen int
	)
	for _, expression := range alert.Rules {
		indexes = append(indexes, expression.AlertIndex)
	}
	ruleMap, err := a.getEnabledAlertRulesByScopeAndIndices(i18n.LanguageCodes{}, alert.AlertScope, alert.AlertScopeID, indexes)
	if err != nil {
		return err
	}
	expressionMap, err := a.getAlertExpressionsMapByAlertID(alertID)
	if err != nil {
		return err
	}
	saveExpressionIDs := make(map[uint64]bool)
	for _, item := range alert.Rules {
		rule, ok := ruleMap[item.AlertIndex]
		if !ok || rule.AlertScope != alert.AlertScope {
			return invalidParameter("rule %s is not scope: %s", rule.AlertIndex.Key, alert.AlertScope)
		}
		expression, err := item.ToModel(alert, rule)
		if err != nil {
			return err
		}
		// if the expression is exist modify it else create
		if _, ok := expressionMap[expression.ID]; ok {
			if err := tx.AlertExpression.Update(expression); err != nil {
				return err
			}
			saveExpressionIDs[item.ID] = true
		} else {
			expression.ID = 0
			if err := tx.AlertExpression.Insert(expression); err != nil {
				return err
			}
		}
		expressionLen++
	}
	if expressionLen == 0 {
		return errors.New("expression is not valid")
	}
	// delete exist expression
	var deleteExpressionIDs []uint64
	for expressionID := range expressionMap {
		if _, ok := saveExpressionIDs[expressionID]; !ok {
			deleteExpressionIDs = append(deleteExpressionIDs, expressionID)
		}
	}
	if err := tx.AlertExpression.DeleteByIDs(deleteExpressionIDs); err != nil {
		return err
	}

	// modify alert notify
	var (
		silence   *AlertNotifySilence
		notifyLen int
	)
	notifyMap, err := a.getAlertNotifysMapByAlertID(alertID)
	if err != nil {
		return err
	}
	saveNotifyIDs := make(map[uint64]bool)
	for _, item := range alert.Notifies {
		silence = item.Silence
		alertNotify := item.ToModel(alert, a.silencePolicies)
		if alertNotify == nil {
			continue
		}
		// if the expression is exist modify it else create
		if _, ok := notifyMap[alertNotify.ID]; ok {
			saveNotifyIDs[alertNotify.ID] = true
			if err := tx.AlertNotify.Update(alertNotify); err != nil {
				return err
			}
		} else {
			var notifyID uint64
			for _, notify := range notifyMap {
				if match := a.compareNotify(alertNotify, notify); match {
					notifyID = notify.ID
					break
				}
			}
			if notifyID != 0 {
				alertNotify.ID = notifyID
				saveNotifyIDs[notifyID] = true
				if err := tx.AlertNotify.Update(alertNotify); err != nil {
					return err
				}
			} else {
				alertNotify.ID = 0
				if err := tx.AlertNotify.Insert(alertNotify); err != nil {
					return err
				}
			}
		}
		notifyLen++
	}
	if notifyLen == 0 {
		return errors.New("notify is not valid")
	}
	// delete exist notify
	var (
		deleteNotifyIDs []uint64
		hasTicket       bool
	)
	for notifyID, notify := range notifyMap {
		if notifyType, ok := utils.GetMapValueString(notify.NotifyTarget, "type"); ok && notifyType == "ticket" {
			hasTicket = true
			// modify ticket notify silence
			t := convertMillisecondByUnit(silence.Value, silence.Unit)
			notify.Silence = t
			if err := tx.AlertNotify.Update(notify); err != nil {
				return err
			}
			continue
		}
		if _, ok := saveNotifyIDs[notifyID]; !ok {
			deleteNotifyIDs = append(deleteNotifyIDs, notifyID)
		}
	}
	if !hasTicket {
		// create ticket alert expression
		notify := a.newTicketAlertNotify(alertID, silence)
		if err := tx.AlertNotify.Insert(notify); err != nil {
			return err
		}
	}
	if err := tx.AlertNotify.DeleteByIDs(deleteNotifyIDs); err != nil {
		return err
	}
	return nil
}

func (a *Adapt) getAlertExpressionsMapByAlertID(alertID uint64) (map[uint64]*db.AlertExpression, error) {
	if alertID == 0 {
		return nil, nil
	}
	expressions, err := a.db.AlertExpression.QueryByAlertIDs([]uint64{alertID})
	if err != nil {
		return nil, err
	}
	expressionsMap := make(map[uint64]*db.AlertExpression)
	for _, expression := range expressions {
		expressionsMap[expression.ID] = expression
	}
	return expressionsMap, nil
}

func (a *Adapt) getAlertNotifysMapByAlertID(alertID uint64) (map[uint64]*db.AlertNotify, error) {
	if alertID == 0 {
		return nil, nil
	}
	notifies, err := a.db.AlertNotify.QueryByAlertIDs([]uint64{alertID})
	if err != nil {
		return nil, err
	}
	notifys := make(map[uint64]*db.AlertNotify)
	for _, notify := range notifies {
		notifys[notify.ID] = notify
	}
	return notifys, nil
}

func (*Adapt) compareNotify(a, b *db.AlertNotify) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.NotifyTarget == nil || b.NotifyTarget == nil {
		return false
	}
	notifyType, ok := utils.GetMapValueString(a.NotifyTarget, "type")
	if !ok {
		return false
	}
	if notifyType == "notify_group" {
		bNotifyType, _ := utils.GetMapValueString(b.NotifyTarget, "type")
		if notifyType != bNotifyType {
			return false
		}
		aGroupID, _ := utils.GetMapValueInt64(a.NotifyTarget, "group_id")
		bGroupID, _ := utils.GetMapValueInt64(b.NotifyTarget, "group_id")
		if aGroupID != bGroupID {
			return false
		}
		aGroupType, _ := utils.GetMapValueString(a.NotifyTarget, "group_type")
		bGroupType, _ := utils.GetMapValueString(b.NotifyTarget, "group_type")
		if aGroupType != bGroupType {
			return false
		}
	} else if notifyType == "dingding" {
		aAddr, _ := utils.GetMapValueString(a.NotifyTarget, "dingding_url")
		bAddr, _ := utils.GetMapValueString(b.NotifyTarget, "dingding_url")
		if aAddr != bAddr {
			return false
		}
	}
	return true
}

// DeleteAlert .
func (a *Adapt) DeleteAlert(id uint64) (err error) {
	tx := a.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			a.l.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = a.db.Alert.DeleteByID(id)
	if err != nil {
		return err
	}
	err = a.db.AlertNotify.DeleteByAlertID(id)
	if err != nil {
		return err
	}
	err = a.db.AlertExpression.DeleteByAlertID(id)
	if err != nil {
		return err
	}
	return nil
}

// DeleteOrgAlert .
func (a *Adapt) DeleteOrgAlert(id uint64, orgID string) (err error) {
	tx := a.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			a.l.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	alert, err := a.db.Alert.GetByID(id)
	if err != nil {
		return err
	}
	if alert == nil {
		return nil
	}
	if alert.AlertScope != "org" || alert.AlertScopeID != orgID {
		return fmt.Errorf("permission denied")
	}
	err = a.db.Alert.DeleteByID(id)
	if err != nil {
		return err
	}
	err = a.db.AlertNotify.DeleteByAlertID(id)
	if err != nil {
		return err
	}
	err = a.db.AlertExpression.DeleteByAlertID(id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateAlertEnable .
func (a *Adapt) UpdateAlertEnable(id uint64, enable bool) (err error) {
	tx := a.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			a.l.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	// close alert
	if err := a.db.Alert.UpdateEnable(id, enable); err != nil {
		return err
	}
	// close alert expression
	if err := a.db.AlertExpression.UpdateEnableByAlertID(id, enable); err != nil {
		return err
	}
	// close alert notify
	if err := a.db.AlertNotify.UpdateEnableByAlertID(id, enable); err != nil {
		return err
	}
	return nil
}

// UpdateOrgAlertEnable .
func (a *Adapt) UpdateOrgAlertEnable(id uint64, enable bool, orgID string) (err error) {
	tx := a.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			a.l.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	alert, err := a.db.Alert.GetByID(id)
	if err != nil {
		return err
	}
	if alert == nil {
		return nil
	}
	if alert.AlertScope != "org" || alert.AlertScopeID != orgID {
		return fmt.Errorf("permission denied")
	}
	// close alert
	if err := a.db.Alert.UpdateEnable(id, enable); err != nil {
		return err
	}
	// close alert expression
	if err := a.db.AlertExpression.UpdateEnableByAlertID(id, enable); err != nil {
		return err
	}
	// close alert notify
	if err := a.db.AlertNotify.UpdateEnableByAlertID(id, enable); err != nil {
		return err
	}
	return nil
}
