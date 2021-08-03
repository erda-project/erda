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

	"github.com/erda-project/erda-proto-go/core/dicehub/template/pb"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type DicePipelineTemplate struct {
	dbengine.BaseModel
	Name           string `json:"name" gorm:"type:varchar(255)"`
	LogoUrl        string `json:"logoUrl" gorm:"type:varchar(255)"`
	Desc           string `json:"desc" gorm:"type:varchar(255)"`
	ScopeType      string `json:"scope_type" gorm:"type:varchar(255)"`
	ScopeId        string `json:"scope_id" gorm:"type:bigint(20)"`
	DefaultVersion string `json:"default_version" gorm:"type:varchar(255)"`
}

func (ext *DicePipelineTemplate) ToApiData() *pb.PipelineTemplate {
	return &pb.PipelineTemplate{
		Id:        ext.ID,
		Name:      ext.Name,
		Desc:      ext.Desc,
		LogoUrl:   ext.LogoUrl,
		CreatedAt: timestamppb.New(ext.CreatedAt),
		UpdatedAt: timestamppb.New(ext.UpdatedAt),
		ScopeID:   ext.ScopeId,
		ScopeType: ext.ScopeType,
		Version:   ext.DefaultVersion,
	}
}

type DicePipelineTemplateVersion struct {
	dbengine.BaseModel
	TemplateId uint64 `json:"template_id"`
	Name       string `json:"name" gorm:"type:varchar(255);"`
	Version    string `json:"version" gorm:"type:varchar(128);"`
	Spec       string `json:"spec" gorm:"type:text"`
	Readme     string `json:"readme" gorm:"type:longtext"`
}

func (ext *DicePipelineTemplateVersion) ToApiData() *pb.PipelineTemplateVersion {
	return &pb.PipelineTemplateVersion{
		Id:         ext.ID,
		Name:       ext.Name,
		TemplateId: ext.TemplateId,
		Version:    ext.Version,
		Spec:       ext.Spec,
		Readme:     ext.Readme,
		CreatedAt:  timestamppb.New(ext.CreatedAt),
		UpdatedAt:  timestamppb.New(ext.UpdatedAt),
	}
}
