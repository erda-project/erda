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
	"reflect"
	"time"

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// tables name
const (
	TableProject = "sp_project"
	TableMetric  = "sp_metric"
)

// Project .
type Project struct {
	ID          int64     `gorm:"column:id;primary_key"`
	Identity    string    `gorm:"column:identity"`
	Name        string    `gorm:"column:config"`
	Description string    `gorm:"column:description"`
	Ats         string    `gorm:"column:ats"`
	Callback    string    `gorm:"column:callback"`
	ProjectID   int64     `gorm:"column:project_id"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
	IsDeleted   string    `gorm:"column:is_deleted"`
}

// TableName .
func (Project) TableName() string { return TableProject }

var projectFieldColumns = gormutil.GetFieldToColumnMap(reflect.TypeOf(Project{}))

// Metric .
type Metric struct {
	ID         int64     `gorm:"column:id;primary_key"`
	ProjectID  int64     `gorm:"column:project_id"`
	ServiceID  int64     `gorm:"column:service_id"`
	Name       string    `gorm:"column:name"`
	URL        string    `gorm:"column:url"`
	Mode       string    `gorm:"column:mode"`
	Extra      string    `gorm:"column:extra"`
	AccountID  int64     `gorm:"column:account_id"`
	Status     int64     `gorm:"column:status"`
	Env        string    `gorm:"column:env"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	IsDeleted  string    `gorm:"column:is_deleted"`
}

// TableName .
func (Metric) TableName() string { return TableMetric }

var metricFieldColumns = gormutil.GetFieldToColumnMap(reflect.TypeOf(Metric{}))
