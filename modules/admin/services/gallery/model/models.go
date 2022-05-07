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

// Opus is the model `erda_gallery_opus`
type Opus struct {
	Model
	Common

	Level            string `gorm:"level"`
	Type             string `gorm:"type"`
	Name             string `gorm:"name"`
	DisplayName      string `gorm:"display_name"`
	Catalog          string `gorm:"catalog"`
	DefaultVersionID string `gorm:"default_version_id"`
	LatestVersionID  string `gorm:"latest_version_id"`
}

func (Opus) TableName() string {
	return "erda_gallery_opus"
}

type OpusVersion struct {
	Model
	Common

	OpusID        string `gorm:"opus_id"`
	Version       string `gorm:"version"`
	Summary       string `gorm:"summary"`
	Labels        string `gorm:"labels"`
	LogoURL       string `gorm:"logo_url"`
	CheckValidURL string `gorm:"check_valid_url"`
	IsValid       bool   `gorm:"is_valid"`
}

func (OpusVersion) TableName() string {
	return "erda_gallery_opus_version"
}

type OpusPresentation struct {
	Model
	Common

	OpusID    string `gorm:"opus_id"`
	VersionID string `gorm:"version_id"`

	Ref             string `gorm:"ref"`
	Desc            string `gorm:"desc"`
	ContactName     string `gorm:"contact_name"`
	ContactURL      string `gorm:"contact_url"`
	ContactEmail    string `gorm:"contact_email"`
	IsOpenSourced   bool   `gorm:"is_open_sourced"`
	OpensourceURL   string `gorm:"opensource_url"`
	LicenceName     string `gorm:"licence_name"`
	LicenceURL      string `gorm:"licence_url"`
	HomepageName    string `gorm:"homepage_name"`
	HomepageURL     string `gorm:"homepage_url"`
	HomepageLogoURL string `gorm:"homepage_logo_url"`
	IsDownloadable  bool   `gorm:"is_downloadable"`
	DownloadURL     string `gorm:"download_url"`
	Parameters      string `gorm:"parameters"`
	Forms           string `gorm:"forms"`
	I18n            string `gorm:"i18n"`
}

func (OpusPresentation) TableName() string {
	return "erda_gallery_opus_presentation"
}

type OpusReadme struct {
	Model
	Common

	OpusID    string `gorm:"opus_id"`
	VersionID string `gorm:"version_id"`

	Lang     string `gorm:"lang"`
	LangName string `gorm:"lang_name"`
	Text     string `gorm:"text"`
}

func (OpusReadme) TableName() string {
	return "erda_gallery_opus_readme"
}

type OpusExtra struct {
	Model
	Common

	OpusID    string `gorm:"opus_id"`
	VersionID string `gorm:"version_id"`

	Extra string `gorm:"extra"`
}

func (OpusExtra) TableName() string {
	return "erda_gallery_opus_extra"
}

type OpusInstallation struct {
	Model
	Common

	OpusID    string `gorm:"opus_id"`
	VersionID string `gorm:"version_id"`

	Installer string `gorm:"installer"`
	Spec      string `gorm:"spec"`
}

func (OpusInstallation) TableName() string {
	return "erda_gallery_opus_installation"
}
