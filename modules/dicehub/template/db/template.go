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

import "github.com/jinzhu/gorm"

// TemplateDB .
type TemplateDB struct {
	*gorm.DB
}

func (client *TemplateDB) CreatePipelineTemplate(template *DicePipelineTemplate) error {
	err := client.Create(template).Error
	return err
}

func (client *TemplateDB) UpdatePipelineTemplate(template *DicePipelineTemplate) error {
	err := client.Model(template).Save(template).Error
	return err
}

func (client *TemplateDB) GetPipelineTemplate(name string, scopeType string, scopeId string) (*DicePipelineTemplate, error) {
	var result DicePipelineTemplate
	err := client.Model(&DicePipelineTemplate{}).Where("name = ? and scope_type = ? and scope_id = ?", name, scopeType, scopeId).Find(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (client *TemplateDB) GetPipelineTemplateVersion(version string, templateId uint64) (*DicePipelineTemplateVersion, error) {
	var result DicePipelineTemplateVersion
	err := client.Model(&DicePipelineTemplate{}).Where(" template_id = ? and version = ? ", templateId, version).Find(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (client *TemplateDB) CreatePipelineTemplateVersion(version *DicePipelineTemplateVersion) error {
	err := client.Create(version).Error
	return err
}

func (client *TemplateDB) UpdatePipelineTemplateVersion(version *DicePipelineTemplateVersion) error {
	err := client.Model(version).Save(version).Error
	return err
}

func (client *TemplateDB) QueryByPipelineTemplates(template *DicePipelineTemplate, pageSize int32, pageNo int32) ([]DicePipelineTemplate, int32, error) {

	var result []DicePipelineTemplate
	var total int32

	query := client.Model(&DicePipelineTemplate{})
	if template != nil {
		if template.Name != "" {
			query = query.Where(" name like ? ", "%"+template.Name+"%")
		}
		if template.ScopeType != "" {
			query = query.Where(" scope_type = ? ", template.ScopeType)
		}
		if template.ScopeId != "" {
			query = query.Where(" scope_id = ? ", template.ScopeId)
		}
	}

	if pageNo > 0 && pageSize > 0 {

		if query.Count(&total).Error == gorm.ErrRecordNotFound {
			return nil, total, nil
		}

		err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&result).Error

		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}

		return result, total, nil
	} else {

		err := query.Find(&result).Error
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}

		return result, 0, err
	}
}

func (client *TemplateDB) QueryPipelineTemplateVersions(version *DicePipelineTemplateVersion) ([]DicePipelineTemplateVersion, error) {
	query := client.Model(&DicePipelineTemplateVersion{})
	if version != nil {
		if version.TemplateId > 0 {
			query = query.Where(" template_id = ? ", version.TemplateId)
		}

		if version.Name != "" {
			query = query.Where(" name = ? ", version.Name)
		}
	}

	var result []DicePipelineTemplateVersion
	err := query.Find(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return result, err
}
