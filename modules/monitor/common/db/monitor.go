// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package db

import "github.com/jinzhu/gorm"

type MonitorDb struct {
	*gorm.DB
}

func (db *MonitorDb) SelectProjectIdByTk(tk string) (string, error) {
	var monitor Monitor
	err := db.
		Select("*").
		Where("sp_monitor.terminus_key = ?", tk).
		Find(&monitor).
		Order("created", false).
		Limit(1).
		Error
	return monitor.ProjectId, err
}

func (db *MonitorDb) GetInstanceByTk(tk string) (Monitor, error) {
	var monitor Monitor
	err := db.Table("sp_monitor").
		Select("*").
		Where("terminus_key = ?", tk).
		Find(&monitor).
		Order("created DESC").
		Limit(1).
		Error
	return monitor, err
}

func (db *MonitorDb) GetTkByProjectIdAndWorkspace(projectId string, workspace string) (string, error) {
	var monitor Monitor
	err := db.Table("sp_monitor").
		Select("terminus_key").
		Where("project_id = ?", projectId).Where("workspace = ?", workspace).
		Find(&monitor).
		Order("id DESC").
		Limit(1).
		Error
	return monitor.TerminusKey, err
}
