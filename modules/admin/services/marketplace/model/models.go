// Copyright (c) 2022 Terminus, Inc.
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

package model

import (
	"time"

	"github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
)

type Model struct {
	ID        fields.UUID      `gorm:"id"`
	CreatedAt time.Time        `gorm:"created_at"`
	UpdatedAt time.Time        `gorm:"updated_at"`
	DeletedAt fields.DeletedAt `gorm:"deleted_at"`
}

type Common struct {
	OrgID     uint32 `gorm:"org_id"`
	OrgName   string `gorm:"org_name"`
	CreatorID string `gorm:"creator_id"`
	UpdaterID string `gorm:"updater_id"`
}

// Extension is the model `dice_extension_version`
type Extension struct {
	ID          string    `gorm:"id" json:"id"`
	CreatedAt   time.Time `gorm:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"updated_at" json:"updatedAt"`
	Type        string    `gorm:"type" json:"type"`
	Name        string    `gorm:"name" json:"name"`
	Category    string    `gorm:"category" json:"category"`
	DisplayName string    `gorm:"display_name" json:"displayName"`
	LogoURL     string    `gorm:"logo_url" json:"logoURL"`
	Desc        string    `gorm:"desc" json:"desc"`
	Public      bool      `gorm:"public" json:"public"`
	Labels      string    `gorm:"labels" json:"labels"`
}

func (Extension) TableName() string {
	return "dice_extension"
}

// ExtensionVersion is the model `dice_extension_version`
type ExtensionVersion struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExtensionID uint64    `json:"extension_id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Spec        string    `json:"spec"`
	Dice        string    `json:"dice"`
	Swagger     string    `json:"swagger"`
	Readme      string    `json:"readme"`
	Public      bool      `json:"public"`
	IsDefault   bool      `json:"is_default"`
}

func (ExtensionVersion) TableName() string {
	return "dice_extension_version"
}

// MarketplaceGalleryArtifacts is the model `erda_marketplace_gallery_artifacts`
type MarketplaceGalleryArtifacts struct {
	Model
	Common

	ReleaseID   string `json:"releaseID" gorm:"release_id"`
	Name        string `json:"name" gorm:"name"`
	DisplayName string `json:"displayName" gorm:"display_name"`
	Version     string `json:"version" gorm:"version"`
	Type        string `json:"type" gorm:"type"`
	Spec        string `json:"spec" gorm:"spec"`
	Changelog   string `json:"changelog" gorm:"changelog"`
	IsDefault   bool   `json:"isDefault" gorm:"is_default"`
}

func (MarketplaceGalleryArtifacts) TableName() string {
	return "erda_marketplace_gallery_artifacts"
}
