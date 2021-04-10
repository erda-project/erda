package adapt

import (
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/cql"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/monitor/utils"
)

// FromModel .
func (a *AlertRecord) FromModel(m *db.AlertRecord) *AlertRecord {
	a.GroupID = m.GroupID
	a.Scope = m.Scope
	a.ScopeKey = m.ScopeKey
	a.AlertGroup = m.AlertGroup
	a.Title = m.Title
	a.AlertState = m.AlertState
	a.AlertType = m.AlertType
	a.AlertIndex = m.AlertIndex
	a.ExpressionKey = m.ExpressionKey
	a.AlertID = m.AlertID
	a.AlertName = m.AlertName
	a.RuleID = m.RuleID
	a.IssueID = m.IssueID
	a.HandleState = m.HandleState
	a.HandlerID = m.HandlerID
	a.AlertTime = utils.ConvertTimeToMS(m.AlertTime)
	a.HandleTime = utils.ConvertTimeToMS(m.HandleTime)
	a.CreateTime = utils.ConvertTimeToMS(m.CreateTime)
	a.UpdateTime = utils.ConvertTimeToMS(m.UpdateTime)
	return a
}

// FromModel .
func (a *AlertHistory) FromModel(m *cql.AlertHistory) *AlertHistory {
	a.GroupID = m.GroupID
	a.Timestamp = m.Timestamp
	a.AlertState = m.AlertState
	a.Title = m.Title
	a.Content = m.Content
	a.DisplayURL = m.DisplayURL
	return a
}
