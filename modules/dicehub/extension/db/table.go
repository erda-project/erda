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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type Extension struct {
	dbengine.BaseModel
	Type        string `json:"type" gorm:"type:varchar(128)"` // type addon|action
	Name        string `json:"name" grom:"type:varchar(128);unique_index"`
	Category    string `json:"category" grom:"type:varchar(128)"`
	DisplayName string `json:"displayName" grom:"type:varchar(128)"`
	LogoUrl     string `json:"logoUrl" grom:"type:varchar(128)"`
	Desc        string `json:"desc" grom:"type:text"`
	Labels      string `json:"labels" grom:"type:labels"`
	Public      bool   `json:"public" ` // Whether to display in the service market
}

func (Extension) TableName() string {
	return "dice_extension"
}

func (ext *Extension) ToApiData() *pb.Extension {
	return &pb.Extension{
		Id:          ext.ID,
		Name:        ext.Name,
		Desc:        ext.Desc,
		Type:        ext.Type,
		DisplayName: ext.DisplayName,
		Category:    ext.Category,
		LogoUrl:     ext.LogoUrl,
		Public:      ext.Public,
		CreatedAt:   timestamppb.New(ext.CreatedAt),
		UpdatedAt:   timestamppb.New(ext.UpdatedAt),
	}
}

type ExtensionVersion struct {
	dbengine.BaseModel
	ExtensionId uint64 `json:"extensionId"`
	Name        string `gorm:"type:varchar(128);index:idx_name"`
	Version     string `json:"version" gorm:"type:varchar(128);index:idx_version"`
	Spec        string `gorm:"type:text"`
	Dice        string `gorm:"type:text"`
	Swagger     string `gorm:"type:longtext"`
	Readme      string `gorm:"type:longtext"`
	Public      bool   `json:"public"`
	IsDefault   bool   `json:"is_default"`
}

func (ExtensionVersion) TableName() string {
	return "dice_extension_version"
}
