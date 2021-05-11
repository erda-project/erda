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

package alert_record

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
)

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	alertRecord := &adapt.AlertRecord{}
	if err := json.Unmarshal(value, alertRecord); err != nil {
		return err
	}
	record := &db.AlertRecord{}
	alertRecord.ToModel(record)
	err := p.mysql.Exec("INSERT INTO `sp_alert_record`"+
		"(`group_id`, `scope`, `scope_key`, `alert_group`, `title`, `alert_state`, `alert_type`, `alert_index`, "+
		"`expression_key`, `alert_id`, `alert_name`, `rule_id`, `alert_time`) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) "+
		"ON DUPLICATE KEY UPDATE "+
		"`alert_state` = ?, `alert_type` = ?, `alert_index` = ?, `alert_name` = ?, `alert_time` = ?",
		record.GroupID, record.Scope, record.ScopeKey, record.AlertGroup, record.Title, record.AlertState, record.AlertType, record.AlertIndex,
		record.ExpressionKey, record.AlertID, record.AlertName, record.RuleID, record.AlertTime,
		record.AlertState, record.AlertType, record.AlertIndex, record.AlertName, record.AlertTime).Error
	fmt.Println("save alert record is succ")
	return err
}
