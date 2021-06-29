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

package db

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// MetricDB .
type MetricDB struct {
	*gorm.DB
}

func (db *MetricDB) query() *gorm.DB {
	return db.Table(TableMetric).Where("`is_deleted`=?", "N")
}

func (db *MetricDB) GetByFields(fields map[string]interface{}) (*Metric, error) {
	query, err := gormutil.GetQueryFilterByFields(db.query(), metricFieldColumns, fields)
	if err != nil {
		return nil, err
	}
	var list []*Metric
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	return list[0], nil
}

func (db *MetricDB) GetByID(id int64) (*Metric, error) {
	return db.GetByFields(map[string]interface{}{
		"ID": id,
	})
}

func (db *MetricDB) ListIDs() ([]int64, error) {
	var list []int64
	if err := db.query().Select("`id`").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (db *MetricDB) ListByIDs(ids ...int64) ([]*Metric, error) {
	var list []*Metric
	if err := db.query().Where("`id` IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (db *MetricDB) Create(m *Metric) error {
	m.ID = 0
	m.IsDeleted = "N"
	return db.Table(TableMetric).Create(m).Error
}

func (db *MetricDB) Update(m *Metric) error {
	m.UpdateTime = time.Now()
	return db.Table(TableMetric).Save(m).Error
}

func (db *MetricDB) Delete(id int64) error {
	return db.Table(TableMetric).Where("`id`=?", id).Updates(map[string]interface{}{
		"is_deleted":  "Y",
		"update_time": time.Now(),
	}).Error
}

func (db *MetricDB) ListByProjectID(projectID int64) ([]*Metric, error) {
	var list []*Metric
	if err := db.query().Where("`project_id`=?", projectID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
