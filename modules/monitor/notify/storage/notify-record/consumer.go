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

package notify_record

import (
	"encoding/json"
	"time"

	"github.com/erda-project/erda/modules/monitor/notify/template/model"
	"github.com/erda-project/erda/modules/monitor/notify/template/query"
)

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	notifyRecord := model.NotifyRecord{}
	if err := json.Unmarshal(value, notifyRecord); err != nil {
		return err
	}
	record := query.ToNotifyRecord(&notifyRecord)
	err := p.mysql.Exec("INSERT INTO `sp_notify_record`"+
		"(`notify_id`,`notify_name`,`scope_type`,`scope_id`,`group_id`,`notify_group`,"+
		"`title`,`notify_time`)"+
		"VALUES (?,?,?,?,?,?,?,?)"+
		"ON DUPLICATE KEY UPDATE"+
		"`notify_id` = ?,`notify_name` = ?,`notify_time` = ?,`title` = ?",
		record.NotifyId, record.NotifyName, record.ScopeType, record.ScopeId, record.GroupId, record.NotifyGroup,
		record.Title, record.NotifyTime, record.NotifyId, record.NotifyName, record.NotifyTime, record.Title).Error
	return err
}
