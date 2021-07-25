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

import "github.com/erda-project/erda/pkg/database/dbengine"

type Image struct {
	dbengine.BaseModel
	ReleaseID string `json:"releaseId" gorm:"index:idx_release_id"`       // release
	ImageName string `json:"imageName" gorm:"type:varchar(128);not null"` // image name
	ImageTag  string `json:"imageTag" gorm:"type:varchar(64)"`            // image tag
	Image     string `json:"image" gorm:"not null"`                       // image addr
}

// Set table name
func (Image) TableName() string {
	return "ps_images"
}
