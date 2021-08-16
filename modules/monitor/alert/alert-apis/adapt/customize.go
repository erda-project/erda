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
	"fmt"
	"sort"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/utils"
)

// FieldMeta .
type FieldMeta struct {
	Field    *DisplayKey `json:"field"`
	DataType string      `json:"dataType"`
}

func (fm *FieldMeta) String() string {
	return fmt.Sprintf("FieldMeta. key: %s, dataType: %s", fm.Field.Key, fm.DataType)
}

// FieldMetaSlice .
type FieldMetaSlice []*FieldMeta

func (s FieldMetaSlice) Len() int {
	return len(s)
}
func (s FieldMetaSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s FieldMetaSlice) Less(i, j int) bool {
	if s[i].Field.Key == s[i].Field.Display {
		if s[j].Field.Key == s[j].Field.Display {
			return s[i].Field.Key < s[j].Field.Key
		} else {
			return false
		}
	} else {
		if s[j].Field.Key == s[j].Field.Display {
			return true
		} else {
			return s[i].Field.Key < s[j].Field.Key
		}
	}
}

// TagMeta .
type TagMeta struct {
	Tag      *DisplayKey `json:"tag"`
	DataType string      `json:"dataType"`
}

func (tm *TagMeta) String() string {
	return fmt.Sprintf("FieldMeta. key: %s, dataType: %s", tm.Tag.Key, tm.DataType)
}

// TagMetaSlice .
type TagMetaSlice []*TagMeta

func (s TagMetaSlice) Len() int {
	return len(s)
}
func (s TagMetaSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s TagMetaSlice) Less(i, j int) bool {
	if s[i].Tag.Key == s[i].Tag.Display {
		if s[j].Tag.Key == s[j].Tag.Display {
			return s[i].Tag.Key < s[j].Tag.Key
		} else {
			return false
		}
	} else {
		if s[j].Tag.Key == s[j].Tag.Display {
			return true
		} else {
			return s[i].Tag.Key < s[j].Tag.Key
		}
	}
}

// MetricMeta .
type MetricMeta struct {
	Name   *DisplayKey  `json:"name"`
	Fields []*FieldMeta `json:"fields"`
	Tags   []*TagMeta   `json:"tags"`
}

// MetricMetaSlice .
type MetricMetaSlice []*MetricMeta

func (s MetricMetaSlice) Len() int {
	return len(s)
}
func (s MetricMetaSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s MetricMetaSlice) Less(i, j int) bool {
	if s[i].Name.Key == s[i].Name.Display {
		if s[j].Name.Key == s[j].Name.Display {
			return s[i].Name.Key < s[j].Name.Key
		}
		return false
	} else {
		if s[j].Name.Key == s[j].Name.Display {
			return true
		}
		return s[i].Name.Key < s[j].Name.Key
	}
}

// CustomizeMetrics .
type CustomizeMetrics struct {
	Metrics           []*MetricMeta `json:"metrics"`
	FunctionOperators []*Operator   `json:"functionOperators"`
	FilterOperators   []*Operator   `json:"filterOperators"`
	Aggregator        []*DisplayKey `json:"aggregator"`
	NotifySample      string        `json:"notifySample"`
}

// CustomizeMetrics .
func (a *Adapt) CustomizeMetrics(lang i18n.LanguageCodes, scope, scopeID string, names []string) (*CustomizeMetrics, error) {
	meta, err := a.metricq.MetricMeta(lang, scope, scopeID, names...)
	if err != nil {
		return nil, err
	}
	var metrics []*MetricMeta
	for _, m := range meta {
		metric := &MetricMeta{
			Name: &DisplayKey{Key: m.Name.Key, Display: m.Name.Name},
		}
		for _, field := range m.Fields {
			metric.Fields = append(metric.Fields, &FieldMeta{
				Field:    &DisplayKey{Key: field.Key, Display: field.Name},
				DataType: field.Type,
			})
		}
		sort.Sort(FieldMetaSlice(metric.Fields))
		for _, tag := range m.Tags {
			metric.Tags = append(metric.Tags, &TagMeta{
				Tag:      &DisplayKey{Key: tag.Key, Display: tag.Name},
				DataType: "string",
			})
		}
		sort.Sort(TagMetaSlice(metric.Tags))
		metrics = append(metrics, metric)
	}
	sort.Sort(MetricMetaSlice(metrics))

	return &CustomizeMetrics{
		Metrics:           metrics,
		FunctionOperators: a.FunctionOperatorKeys(lang),
		FilterOperators:   a.FilterOperatorKeys(lang),
		Aggregator:        a.AggregatorKeys(lang),
	}, nil
}

type (
	// CustomizeAlertRule .
	CustomizeAlertRule struct {
		ID                  uint64                        `json:"id"`
		Name                string                        `json:"name"`
		Metric              string                        `json:"metric"`
		Window              uint64                        `json:"window"`
		Functions           []*CustomizeAlertRuleFunction `json:"functions"`
		Filters             []*CustomizeAlertRuleFilter   `json:"filters"`
		Group               []string                      `json:"group"`
		Outputs             []string                      `json:"outputs"`
		Select              map[string]string             `json:"select"`
		Attributes          map[string]interface{}        `json:"attributes"`
		ActivedMetricGroups []string                      `json:"activedMetricGroups,omitempty"` // for frontend
		CreateTime          int64                         `json:"createTime"`
		UpdateTime          int64                         `json:"updateTime"`
	}
	// CustomizeAlertRuleFunction .
	CustomizeAlertRuleFunction struct {
		Field      string      `json:"field"`
		Alias      string      `json:"alias"`
		Aggregator string      `json:"aggregator"`
		Operator   string      `json:"operator"`
		Value      interface{} `json:"value"`
		DataType   string      `json:"dataType"`
		Unit       string      `json:"unit" translate:"true"`
	}
	// CustomizeAlertRuleFilter .
	CustomizeAlertRuleFilter struct {
		Tag      string      `json:"tag"`
		Operator string      `json:"operator"`
		Value    interface{} `json:"value"`
		DataType string      `json:"dataType"`
	}
	// CustomizeAlertRuleTemplate .
	CustomizeAlertRuleTemplate struct {
		Metric    string                        `json:"metric"`
		Window    uint64                        `json:"window"`
		Functions []*CustomizeAlertRuleFunction `json:"functions"`
		Filters   []*CustomizeAlertRuleFilter   `json:"filters"`
		Group     []string                      `json:"group"`
		Outputs   []string                      `json:"outputs"`
		Select    map[string]string             `json:"select"`
	}
	// CustomizeAlertNotifyTemplate .
	CustomizeAlertNotifyTemplate struct {
		ID         uint64                 `json:"id"`
		Name       string                 `json:"name"`
		Target     string                 `json:"target"`
		Title      string                 `json:"title"`
		Content    string                 `json:"content"`
		Attributes map[string]interface{} `json:"attributes"`
		CreateTime int64                  `json:"createTime"`
		UpdateTime int64                  `json:"updateTime"`
	}
	// CustomizeAlertOverview .
	CustomizeAlertOverview struct {
		ID            uint64   `json:"id"`
		Name          string   `json:"name"`
		Metric        string   `json:"metric"`
		Window        uint64   `json:"window"`
		NotifyTargets []string `json:"notifyTargets"`
		DashboardID   string   `json:"dashboardId,omitempty"`
		Enable        bool     `json:"enable"`
		CreateTime    int64    `json:"createTime"`
		UpdateTime    int64    `json:"updateTime"`
	}
	// CustomizeAlertNotifyTemplates .
	CustomizeAlertNotifyTemplates struct {
		ID         uint64                 `json:"id"`
		Name       string                 `json:"name"`
		Targets    []string               `json:"targets"`
		Title      string                 `json:"title"`
		Content    string                 `json:"content"`
		Attributes map[string]interface{} `json:"attributes"`
		CreateTime int64                  `json:"createTime"`
		UpdateTime int64                  `json:"updateTime"`
	}
	// CustomizeAlertDetail .
	CustomizeAlertDetail struct {
		ID           uint64                           `json:"id"`
		ClusterName  string                           `json:"clusterName"`
		Name         string                           `json:"name"`
		AlertType    string                           `json:"alertType"`
		AlertScope   string                           `json:"alertScope"`
		AlertScopeID string                           `json:"alertScopeId"`
		Enable       bool                             `json:"enable"`
		Attributes   map[string]interface{}           `json:"attributes"`
		Rules        []*CustomizeAlertRule            `json:"rules"`
		Notifies     []*CustomizeAlertNotifyTemplates `json:"notifies"`
		CreateTime   int64                            `json:"createTime"`
		UpdateTime   int64                            `json:"updateTime"`

		// just for frontend
		Lang i18n.LanguageCodes `json:"-"`
	}
)

// customize alert type
const (
	customizeAlertTypeOrg          = "org_customize"
	customizeAlertTypeMicroService = "micro_service_customize"
	applicationIdTag               = "application_id"
	applicationIdValue             = "$application_id"
	clusterNameTag                 = "cluster_name"
	clusterNameValue               = "$cluster_name"
)

// CustomizeAlerts .
func (a *Adapt) CustomizeAlerts(lang i18n.LanguageCodes, scope, scopeID string, pageNo, pageSize int) ([]*CustomizeAlertOverview, int, error) {
	alerts, err := a.db.CustomizeAlert.QueryByScopeAndScopeID(scope, scopeID, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var alertIDs []uint64
	for _, alert := range alerts {
		alertIDs = append(alertIDs, alert.ID)
	}
	// get alert rule
	rulesMap, err := a.getCustomizeAlertRulesByAlertIDs(alertIDs)
	if err != nil {
		return nil, 0, err
	}
	// get alert notify template
	notifyTemplatesMap, err := a.getCustomizeAlertNotifyTemplatesByAlertIDs(alertIDs)
	if err != nil {
		return nil, 0, err
	}
	var list []*CustomizeAlertOverview
	for _, item := range alerts {
		alert := a.newCustomizeAlertOverview(lang, item, rulesMap[item.ID], notifyTemplatesMap[item.ID])
		if alert == nil {
			continue
		}
		list = append(list, alert)
	}
	total, err := a.db.CustomizeAlert.CountByScopeAndScopeID(scope, scopeID)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// return alertID to rules Map
func (a *Adapt) getCustomizeAlertRulesByAlertIDs(id []uint64) (map[uint64][]*CustomizeAlertRule, error) {
	rules, err := a.db.CustomizeAlertRule.QueryByAlertIDs(id)
	if err != nil {
		return nil, err
	}
	alertRules := make(map[uint64][]*CustomizeAlertRule)
	for _, item := range rules {
		rule, err := (&CustomizeAlertRule{}).FromModel(item)
		if err != nil {
			a.l.Errorf("fail to wrap customize rule: %s", err)
			continue
		} else if rule == nil {
			continue
		}
		alertRules[item.CustomizeAlertID] = append(alertRules[item.CustomizeAlertID], rule)
	}
	return alertRules, nil
}

// return alertID to notify template Map
func (a *Adapt) getCustomizeAlertNotifyTemplatesByAlertIDs(id []uint64) (map[uint64][]*CustomizeAlertNotifyTemplate, error) {
	notifyTemplates, err := a.db.CustomizeAlertNotifyTemplate.QueryByAlertIDs(id)
	if err != nil {
		return nil, err
	}
	notifyTemplatesMap := make(map[uint64][]*CustomizeAlertNotifyTemplate)
	for _, item := range notifyTemplates {
		notify := (&CustomizeAlertNotifyTemplate{}).FromModel(item)
		if notify == nil {
			continue
		}
		notifyTemplatesMap[item.CustomizeAlertID] = append(notifyTemplatesMap[item.CustomizeAlertID], notify)
	}
	return notifyTemplatesMap, nil
}

// CustomizeAlert .
func (a *Adapt) CustomizeAlert(id uint64) (*CustomizeAlertDetail, error) {
	alert, err := a.db.CustomizeAlert.GetByID(id)
	if err != nil {
		return nil, err
	}
	return (&CustomizeAlertDetail{}).FromModel(alert), nil
}

// CustomizeAlertDetail .
func (a *Adapt) CustomizeAlertDetail(id uint64) (*CustomizeAlertDetail, error) {
	alert, err := a.db.CustomizeAlert.GetByID(id)
	if err != nil {
		return nil, err
	}
	alertIDs := []uint64{id}
	rulesMap, err := a.getCustomizeAlertRulesByAlertIDs(alertIDs)
	if err != nil {
		return nil, err
	}
	rules := rulesMap[id]
	notifyTemplatesMap, err := a.getCustomizeAlertNotifyTemplatesByAlertIDs(alertIDs)
	if err != nil {
		return nil, err
	}
	notifyTemplates := notifyTemplatesMap[id]
	alertDetail := (&CustomizeAlertDetail{}).FromModelWithDetail(alert, rules, notifyTemplates)
	if alertDetail == nil {
		return nil, nil
	}
	for _, rule := range alertDetail.Rules {
		// filter
		filters := make([]*CustomizeAlertRuleFilter, 0)
		for _, filter := range rule.Filters {
			if filter.Tag == applicationIdTag && filter.Value == applicationIdValue {
				continue
			}
			if filter.Tag == clusterNameTag && filter.Value.(string) == clusterNameValue {
				continue
			}
			if alert.AlertType == customizeAlertTypeMicroService && a.microServiceFilterTags[filter.Tag] {
				continue
			}
			if alert.AlertType == customizeAlertTypeOrg && a.orgFilterTags[filter.Tag] {
				continue
			}
			filters = append(filters, filter)
		}
		rule.Filters = filters
		// filter group
		groups := make([]string, 0)
		for _, group := range rule.Group {
			if alert.AlertType == customizeAlertTypeMicroService && a.microServiceFilterTags[group] {
				continue
			}
			if alert.AlertType == customizeAlertTypeOrg && a.orgFilterTags[group] {
				continue
			}
			groups = append(groups, group)
		}
		rule.Group = groups
	}
	return alertDetail, nil
}

// CreateCustomizeAlert .
func (a *Adapt) CreateCustomizeAlert(alertDetail *CustomizeAlertDetail) (alertID uint64, err error) {
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

	if alert, err := tx.CustomizeAlert.GetByScopeAndScopeIDAndName(alertDetail.AlertScope, alertDetail.AlertScopeID, alertDetail.Name); err != nil {
		return 0, err
	} else if alert != nil {
		return 0, ErrorAlreadyExists
	}
	// create alert
	index, err := a.generateCustomizeAlertIndex()
	if err != nil {
		return 0, err
	}

	// related to the dashboard
	dashboardID, err := NewDashboard(a).CreateChartDashboard(alertDetail)
	if err != nil {
		return 0, err
	}

	alertDetail.ID = 0
	alertDetail.Enable = true
	if alertDetail.Attributes == nil {
		alertDetail.Attributes = make(map[string]interface{})
	}
	alertDetail.Attributes["alert_index"] = index
	alertDetail.Attributes["alert_dashboard_id"] = dashboardID
	alert := alertDetail.ToModel()
	if err := tx.CustomizeAlert.Insert(alert); err != nil {
		return 0, err
	}
	alertDetail.ID = alert.ID

	// create alarm rules, only one rule is allowed
	alertRule := alertDetail.Rules[0]
	if alertRule.Name == "" {
		alertRule.Name = alertDetail.Name
	}
	rule := alertRule.ToModel(alertDetail, index)
	if err := tx.CustomizeAlertRule.Insert(rule); err != nil {
		return 0, err
	}

	// create an alert notification template
	notifyTargetMap := make(map[string]bool)
	for _, notify := range alertDetail.Notifies {
		var notifyTemplate *CustomizeAlertNotifyTemplate
		for _, target := range notify.Targets {
			// Avoid target duplication
			if ok := notifyTargetMap[target]; ok {
				continue
			}
			notifyTargetMap[target] = true

			if notifyTemplate == nil {
				notifyTemplate = &CustomizeAlertNotifyTemplate{
					ID:         notify.ID,
					Name:       notify.Name,
					Title:      notify.Title,
					Content:    notify.Content,
					CreateTime: notify.CreateTime,
					UpdateTime: notify.UpdateTime,
				}
				if notifyTemplate.Name == "" {
					notifyTemplate.Name = alertDetail.Name
				}
			}
			notifyTemplate.Target = target
			notify, err := notifyTemplate.ToModel(alertDetail, index)
			if err != nil {
				return 0, err
			}
			if err := tx.CustomizeAlertNotifyTemplate.Insert(notify); err != nil {
				return 0, err
			}
		}
	}
	return alertDetail.ID, nil
}

func (a *Adapt) generateCustomizeAlertIndex() (string, error) {
	return utils.UUID()
}

// UpdateCustomizeAlert .
func (a *Adapt) UpdateCustomizeAlert(alertDetail *CustomizeAlertDetail) (err error) {
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
	if alertDetail.Name != "" {
		if alert, err := tx.CustomizeAlert.GetByScopeAndScopeIDAndName(alertDetail.AlertScope, alertDetail.AlertScopeID, alertDetail.Name); err != nil {
			return err
		} else if alert != nil && alert.ID != alertDetail.ID {
			return ErrorAlreadyExists
		}
	}

	// modify alert
	alert, err := tx.CustomizeAlert.GetByID(alertDetail.ID)
	if err != nil {
		return err
	}
	if alert == nil {
		return nil
	}
	index, ok := utils.GetMapValueString(alert.Attributes, "alert_index")
	if !ok {
		return fmt.Errorf("no alert index attributes")
	}

	attributes := make(map[string]interface{})
	for k, v := range alert.Attributes {
		attributes[k] = v
	}
	for k, v := range alertDetail.Attributes {
		attributes[k] = v
	}

	attributes["alert_index"] = index
	alertDetail.Attributes = attributes
	if err := tx.CustomizeAlert.Update(alertDetail.ToModel()); err != nil {
		return err
	}
	// modify alert rule
	rules, err := a.getCustomizeAlertRulesMapByAlertID(alertDetail.ID)
	if err != nil {
		return err
	}
	saveRuleIDs := make(map[uint64]bool)
	for _, item := range alertDetail.Rules {
		if item.Name == "" {
			item.Name = alertDetail.Name
		}
		rule := item.ToModel(alertDetail, index)
		// Modify if it exists, add if it doesn't exist
		if _, ok := rules[rule.ID]; ok {
			if err := tx.CustomizeAlertRule.Update(rule); err != nil {
				return err
			}
			saveRuleIDs[rule.ID] = true
		} else {
			rule.ID = 0
			if err := tx.CustomizeAlertRule.Insert(rule); err != nil {
				return err
			}
		}
	}
	// delete exist expression
	var deleteRuleIDs []uint64
	for ruleID := range rules {
		if _, ok := saveRuleIDs[ruleID]; !ok {
			deleteRuleIDs = append(deleteRuleIDs, ruleID)
		}
	}
	if err := tx.CustomizeAlertRule.DeleteByIDs(deleteRuleIDs); err != nil {
		return err
	}

	// modify alert notify template
	notifyMap, err := a.getCustomizeAlertNotifyTemplateByAlertID(alertDetail.ID)
	if err != nil {
		return err
	}
	notifyTargetMap := make(map[string]*CustomizeAlertNotifyTemplate)
	for _, notify := range notifyMap {
		notifyTargetMap[notify.Name+notify.Target] = notify
	}
	saveNotifyIDs := make(map[uint64]bool)
	for _, notifyDTO := range alertDetail.Notifies {
		var templateDTO *CustomizeAlertNotifyTemplate
		for _, target := range notifyDTO.Targets {
			if templateDTO == nil {
				templateDTO = &CustomizeAlertNotifyTemplate{
					ID:         notifyDTO.ID,
					Name:       notifyDTO.Name,
					Title:      notifyDTO.Title,
					Content:    notifyDTO.Content,
					CreateTime: notifyDTO.CreateTime,
					UpdateTime: notifyDTO.UpdateTime,
				}
				if templateDTO.Name == "" {
					templateDTO.Name = alertDetail.Name
				}
			}
			templateDTO.Target = target
			template, err := templateDTO.ToModel(alertDetail, index)
			if err != nil {
				return err
			}
			// Modify if it exists, add if it doesn't exist
			if notify, ok := notifyTargetMap[template.Name+template.Target]; ok {
				template.ID = notify.ID
				if err := tx.CustomizeAlertNotifyTemplate.Update(template); err != nil {
					return err
				}
				saveNotifyIDs[template.ID] = true
			} else {
				template.ID = 0
				if err := tx.CustomizeAlertNotifyTemplate.Insert(template); err != nil {
					return err
				}
			}
		}
	}
	// delete exist notify template
	var deleteNotifyIDs []uint64
	for notifyID := range notifyMap {
		if _, ok := saveNotifyIDs[notifyID]; !ok {
			deleteNotifyIDs = append(deleteNotifyIDs, notifyID)
		}
	}
	if err := tx.CustomizeAlertNotifyTemplate.DeleteByIDs(deleteNotifyIDs); err != nil {
		return err
	}
	return nil
}

// return rules Map
func (a *Adapt) getCustomizeAlertRulesMapByAlertID(id uint64) (map[uint64]*CustomizeAlertRule, error) {
	rules, err := a.db.CustomizeAlertRule.QueryByAlertIDs([]uint64{id})
	if err != nil {
		return nil, err
	}
	rulesMap := make(map[uint64]*CustomizeAlertRule)
	for _, item := range rules {
		rule, err := (&CustomizeAlertRule{}).FromModel(item)
		if err != nil {
			a.l.Errorf("fail to wrap customize rule: %s", err)
			continue
		} else if rule == nil {
			continue
		}
		rulesMap[item.ID] = rule
	}
	return rulesMap, nil
}

func (a *Adapt) getCustomizeAlertNotifyTemplateByAlertID(alertID uint64) (map[uint64]*CustomizeAlertNotifyTemplate, error) {
	notifyTemplates, err := a.db.CustomizeAlertNotifyTemplate.QueryByAlertIDs([]uint64{alertID})
	if err != nil {
		return nil, err
	}
	notifyTemplateMap := make(map[uint64]*CustomizeAlertNotifyTemplate)
	for _, item := range notifyTemplates {
		notify := (&CustomizeAlertNotifyTemplate{}).FromModel(item)
		if notify == nil {
			continue
		}
		notifyTemplateMap[item.ID] = notify
	}
	return notifyTemplateMap, nil
}

// UpdateCustomizeAlertEnable .
func (a *Adapt) UpdateCustomizeAlertEnable(id uint64, enable bool) (err error) {
	tx := a.db.Begin()
	defer func() {
		if exp := recover(); exp != nil {
			a.l.Errorf("panic: %s", exp)
			tx.Rollback()
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	// close alert
	if err := tx.CustomizeAlert.UpdateEnable(id, enable); err != nil {
		return err
	}
	// close alert expression
	if err := tx.CustomizeAlertRule.UpdateEnableByAlertID(id, enable); err != nil {
		return err
	}
	// close alert notify
	if err := tx.CustomizeAlertNotifyTemplate.UpdateEnableByAlertID(id, enable); err != nil {
		return err
	}
	return nil
}

// DeleteCustomizeAlert .
func (a *Adapt) DeleteCustomizeAlert(id uint64) (err error) {
	tx := a.db.Begin()
	defer func() {
		if exp := recover(); exp != nil {
			a.l.Errorf("panic: %s", exp)
			tx.Rollback()
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err := tx.CustomizeAlert.DeleteByID(id); err != nil {
		return err
	}
	if err := tx.CustomizeAlertRule.DeleteByAlertID(id); err != nil {
		return err
	}
	if err := tx.CustomizeAlertNotifyTemplate.DeleteByAlertID(id); err != nil {
		return err
	}
	return nil
}
