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

package db

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
)

// CheckerDB .
type CheckerDB struct {
	*gorm.DB
	ScopeInfoUpdateInterval time.Duration

	cache           map[scopeCacheKey]*scopeInfo
	lastCacheUpdate time.Time
	lock            sync.Mutex
}

type scopeCacheKey struct {
	ProjectID int64
	Env       string
}

// scopeInfo .
type scopeInfo struct {
	ProjectName string `gorm:"column:project_name"`
	ScopeID     string `gorm:"column:scope_id"`
}

// JoinFileds .
type JoinFileds struct {
	ID          int64  `gorm:"column:id"`
	Name        string `gorm:"column:name"`
	Mode        string `gorm:"column:mode"`
	URL         string `gorm:"column:url"`
	ProjectID   int64  `gorm:"column:project_id"`
	ProjectName string `gorm:"column:project_name"`
	Env         string `gorm:"column:env"`
	ScopeID     string `gorm:"column:scope_id"`
	IsDeleted   string `gorm:"column:is_deleted"`
}

// Deleted .
func (f *JoinFileds) Deleted() bool {
	return f.IsDeleted != "N"
}

var joinFieldSelect = strings.Join([]string{
	"sp_metric.id AS `id`",
	"sp_metric.name AS `name`",
	"sp_metric.mode AS `mode`",
	"sp_metric.url AS `url`",
	"sp_project.project_id AS `project_id`",
	// "sp_monitor.project_name AS `project_name`",
	"sp_metric.env AS `env`",
	// "sp_monitor.terminus_key AS `scope_id`",
	"sp_metric.is_deleted AS `is_deleted`",
}, ", ")

func (db *CheckerDB) FullList() (checkers []*pb.Checker, deleted []int64, err error) {
	var list []*JoinFileds
	if err := db.DB.Table(TableMetric).
		Select(joinFieldSelect).
		Joins("LEFT JOIN sp_project ON sp_project.id = sp_metric.project_id").
		// Joins("LEFT JOIN sp_monitor ON sp_monitor.project_id = sp_project.project_id AND sp_monitor.workspace = sp_metric.env").
		Find(&list).Error; err != nil {
		return nil, nil, err
	}
	for _, item := range list {
		item.IsDeleted = strings.ToUpper(item.IsDeleted)
		if item.Deleted() {
			deleted = append(deleted, item.ID)
		}
	}

	db.lock.Lock()
	defer db.lock.Unlock()
	err = db.updateCacheIfNeed(list)
	if err != nil {
		return nil, nil, err
	}
	for _, item := range list {
		if item.Deleted() {
			continue
		}
		scope, _ := db.cache[scopeCacheKey{ProjectID: item.ProjectID, Env: item.Env}]
		if scope == nil {
			continue
		}
		item.ProjectName = scope.ProjectName
		item.ScopeID = scope.ScopeID
		checkers = append(checkers, convertToChecker(item))
	}
	return checkers, deleted, nil
}

func (db *CheckerDB) updateCacheIfNeed(list []*JoinFileds) error {
	now := time.Now()
	if db.cache == nil || time.Now().After(db.lastCacheUpdate.Add(db.ScopeInfoUpdateInterval)) {
		cache := make(map[scopeCacheKey]*scopeInfo)
		for _, item := range list {
			if item.Deleted() {
				continue
			}
			info, err := db.queryScopeInfo(item.ProjectID, item.Env)
			if err != nil {
				return err
			}
			cache[scopeCacheKey{ProjectID: item.ProjectID, Env: item.Env}] = info
		}
		db.cache, db.lastCacheUpdate = cache, now
	} else {
		for _, item := range list {
			key := scopeCacheKey{ProjectID: item.ProjectID, Env: item.Env}
			if _, ok := db.cache[key]; !ok && !item.Deleted() {
				info, err := db.queryScopeInfo(item.ProjectID, item.Env)
				if err != nil {
					return err
				}
				db.cache[key] = info
			}
		}
	}
	return nil
}

func (db *CheckerDB) queryScopeInfo(projectID int64, env string) (*scopeInfo, error) {
	var info scopeInfo
	if err := db.DB.Table("sp_monitor").
		Select("`terminus_key` AS `scope_id`, `project_name`").
		Where("`project_id`=? AND `workspace`=?", strconv.FormatInt(projectID, 10), env).
		Last(&info).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

func convertToChecker(fields *JoinFileds) *pb.Checker {
	return &pb.Checker{
		Id:   fields.ID,
		Name: fields.Name,
		Type: fields.Mode,
		Config: map[string]string{
			"url": fields.URL,
		},
		Tags: map[string]string{
			"_metric_scope":    "micro_service",
			"_metric_scope_id": fields.ScopeID,
			"terminus_key":     fields.ScopeID,
			"project_id":       strconv.FormatInt(fields.ProjectID, 10),
			"project_name":     fields.ProjectName,
			"env":              fields.Env,
			"metric":           strconv.FormatInt(fields.ID, 10),
			"metric_name":      fields.Name,
		},
	}
}

func (db *MetricDB) QueryScopeInfo(projectID int64, env string) (*scopeInfo, error) {
	var info scopeInfo
	if err := db.DB.Table("sp_monitor").
		Select("`terminus_key` AS `scope_id`, `project_name`").
		Where("`project_id`=? AND `workspace`=?", strconv.FormatInt(projectID, 10), env).
		Last(&info).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}
