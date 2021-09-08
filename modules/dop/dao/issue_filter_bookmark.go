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

package dao

import (
	"github.com/google/uuid"
	"time"
)

// IssueFilterBookmark users' bookmark of issue filter
type IssueFilterBookmark struct {
	ID           string `gorm:"primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Name         string
	UserID       string
	ProjectID    string
	PageKey      string // 4 page and each page has 2 variants, thus 8 page for bookmarks
	FilterEntity string
}

func (IssueFilterBookmark) TableName() string {
	return "erda_issue_filter_bookmark"
}

func (db *DBClient) ListIssueFilterBookmarkByUserIDAndProjectID(userID, projectID string) ([]IssueFilterBookmark, error) {
	if userID == "" {
		return nil, nil
	}
	var bms []IssueFilterBookmark
	if err := db.Where("user_id = ?", userID).Where("project_id = ?", projectID).Find(&bms).Error; err != nil {
		return nil, err
	}
	return bms, nil
}

func (db *DBClient) CreateIssueFilterBookmark(bm *IssueFilterBookmark) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	bm.ID = id.String()
	return db.Create(bm).Error
}

func (db *DBClient) DeleteIssueFilterBookmarkByID(id string) error {
	if id == "" {
		return nil
	}
	return db.Delete(&IssueFilterBookmark{ID: id}).Error
}
