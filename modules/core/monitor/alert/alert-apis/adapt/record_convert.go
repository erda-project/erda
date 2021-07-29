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
	"time"

	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/cql"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/db"
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

func ToPBAlertRecord(m *db.AlertRecord) *pb.AlertRecord {
	a := &pb.AlertRecord{}
	a.GroupId = m.GroupID
	a.Scope = m.Scope
	a.ScopeKey = m.ScopeKey
	a.AlertGroup = m.AlertGroup
	a.Title = m.Title
	a.AlertState = m.AlertState
	a.AlertType = m.AlertType
	a.AlertIndex = m.AlertIndex
	a.ExpressionKey = m.ExpressionKey
	a.AlertId = m.AlertID
	a.AlertName = m.AlertName
	a.RuleId = m.RuleID
	a.IssueId = m.IssueID
	a.HandleState = m.HandleState
	a.HandlerId = m.HandlerID
	a.AlertTime = utils.ConvertTimeToMS(m.AlertTime)
	a.HandleTime = utils.ConvertTimeToMS(m.HandleTime)
	a.CreateTime = utils.ConvertTimeToMS(m.CreateTime)
	a.UpdateTime = utils.ConvertTimeToMS(m.UpdateTime)
	return a
}

func (a *AlertRecord) ToModel(m *db.AlertRecord) {
	m.GroupID = a.GroupID
	m.Scope = a.Scope
	m.ScopeKey = a.ScopeKey
	m.AlertGroup = a.AlertGroup
	m.Title = a.Title
	m.AlertState = a.AlertState
	m.AlertType = a.AlertType
	m.AlertIndex = a.AlertIndex
	m.ExpressionKey = a.ExpressionKey
	m.AlertID = a.AlertID
	m.AlertName = a.AlertName
	m.RuleID = a.RuleID
	m.IssueID = a.IssueID
	m.HandleState = a.HandleState
	m.HandlerID = a.HandlerID
	m.AlertTime = time.Unix(a.AlertTime/1000, 0)
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

func ToDBAlertHistory(m *cql.AlertHistory) *pb.AlertHistory {
	a := &pb.AlertHistory{}
	a.GroupId = m.GroupID
	a.Timestamp = m.Timestamp
	a.AlertState = m.AlertState
	a.Title = m.Title
	a.Content = m.Content
	a.DisplayUrl = m.DisplayURL
	return a
}
