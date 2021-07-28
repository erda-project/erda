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

package cql

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

type AlertHistoryCql struct {
	*gocql.Session
}

func (cql *AlertHistoryCql) TableName() string {
	return "alert_history"
}

func (cql *AlertHistoryCql) QueryAlertHistory(groupID string, start, end int64, limit uint) ([]*AlertHistory, error) {
	stmt, names := qb.Select(cql.TableName()).
		Where(
			qb.Eq("group_id"),
			qb.GtNamed("timestamp", "start"),
			qb.LtNamed("timestamp", "end")).
		OrderBy("timestamp", qb.DESC).
		Limit(limit).ToCql()
	var histories []*AlertHistory
	if err := gocqlx.Query(cql.Query(stmt), names).BindMap(qb.M{
		"group_id": groupID,
		"start":    start,
		"end":      end,
	}).SelectRelease(&histories); err != nil {
		return nil, err
	}
	return histories, nil
}
