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

package dbclient

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type Extension struct {
	dbengine.BaseModel
	Type        string `json:"type" gorm:"type:varchar(128)"` // 类型 addon|action
	Name        string `json:"name" grom:"type:varchar(128);unique_index"`
	Category    string `json:"category" grom:"type:varchar(128)"`
	DisplayName string `json:"displayName" grom:"type:varchar(128)"`
	LogoUrl     string `json:"logoUrl" grom:"type:varchar(128)"`
	Desc        string `json:"desc" grom:"type:text"`
	Labels      string `json:"labels" grom:"type:labels"`
	Public      bool   `json:"public" ` //是否在服务市场展示
}

func (Extension) TableName() string {
	return "dice_extension"
}

func (ext *Extension) ToApiData() *apistructs.Extension {
	return &apistructs.Extension{
		ID:          ext.ID,
		Name:        ext.Name,
		Desc:        ext.Desc,
		Type:        ext.Type,
		DisplayName: ext.DisplayName,
		Category:    ext.Category,
		LogoUrl:     ext.LogoUrl,
		Public:      ext.Public,
		CreatedAt:   ext.CreatedAt,
		UpdatedAt:   ext.UpdatedAt,
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

func (ext *ExtensionVersion) ToApiData(typ string, yamlFormat bool) *apistructs.ExtensionVersion {
	if yamlFormat {
		return &apistructs.ExtensionVersion{
			Name:      ext.Name,
			Type:      typ,
			Version:   ext.Version,
			Dice:      ext.Dice,
			Spec:      ext.Spec,
			Swagger:   ext.Swagger,
			Readme:    ext.Readme,
			CreatedAt: ext.CreatedAt,
			UpdatedAt: ext.UpdatedAt,
			IsDefault: ext.IsDefault,
			Public:    ext.Public,
		}
	} else {
		diceData, _ := yaml.YAMLToJSON([]byte(ext.Dice))
		specData, _ := yaml.YAMLToJSON([]byte(ext.Spec))
		swaggerData, _ := yaml.YAMLToJSON([]byte(ext.Swagger))
		var diceJson interface{}
		var specJson interface{}
		var swaggerJson interface{}
		json.Unmarshal(diceData, &diceJson)
		json.Unmarshal(specData, &specJson)
		json.Unmarshal(swaggerData, &swaggerJson)
		return &apistructs.ExtensionVersion{
			Name:      ext.Name,
			Type:      typ,
			Version:   ext.Version,
			Dice:      diceJson,
			Spec:      specJson,
			Swagger:   swaggerJson,
			Readme:    ext.Readme,
			CreatedAt: ext.CreatedAt,
			UpdatedAt: ext.UpdatedAt,
			IsDefault: ext.IsDefault,
			Public:    ext.Public,
		}
	}
}

func (client *DBClient) CreateExtension(extension *Extension) error {
	var cnt int64
	client.Model(&Extension{}).Where("name = ?", extension.Name).Count(&cnt)
	if cnt == 0 {
		err := client.Create(extension).Error
		return err
	} else {
		return errors.New("name already exist")
	}
}

func (client *DBClient) QueryExtensions(all string, typ string, labels string) ([]Extension, error) {
	var result []Extension
	query := client.Model(&Extension{})

	// 不显式指定all=true,只返回public的数据
	if all != "true" {
		query = query.Where("public = ?", true)
	}

	if typ != "" {
		query = query.Where("type = ?", typ)
	}

	if labels != "" {
		labelPairs := strings.Split(labels, ",")
		for _, pair := range labelPairs {
			if strings.LastIndex(pair, "^") == 0 && len(pair) > 1 {
				query = query.Where("labels not like ?", "%"+pair[1:]+"%")
			} else {
				query = query.Where("labels like ?", "%"+pair+"%")
			}

		}
	}
	err := query.Find(&result).Error
	return result, err
}

func (client *DBClient) GetExtension(name string) (*Extension, error) {
	var result Extension
	err := client.Model(&Extension{}).Where("name = ?", name).Find(&result).Error
	return &result, err
}

func (client *DBClient) DeleteExtension(name string) error {
	return client.Where("name = ?", name).Delete(&Extension{}).Error
}

func (client *DBClient) GetExtensionVersion(name string, version string) (*ExtensionVersion, error) {
	var result ExtensionVersion
	err := client.Model(&ExtensionVersion{}).
		Where("name = ? ", name).
		Where("version = ?", version).
		Find(&result).Error
	return &result, err
}

func (client *DBClient) GetExtensionDefaultVersion(name string) (*ExtensionVersion, error) {
	var result ExtensionVersion
	err := client.Model(&ExtensionVersion{}).
		Where("name = ? ", name).
		Where("is_default = ? ", true).
		Limit(1).
		Find(&result).Error
	//没有默认,找最新更新并且是public的
	if err == gorm.ErrRecordNotFound {
		err = client.Model(&ExtensionVersion{}).
			Where("name = ? ", name).
			Where("public = ? ", true).
			Order("version desc").
			Limit(1).
			Find(&result).Error
	}
	return &result, err
}

func (client *DBClient) SetUnDefaultVersion(name string) error {
	return client.Model(&ExtensionVersion{}).
		Where("is_default = ?", true).
		Where("name = ?", name).
		Update("is_default", false).Error
}

func (client *DBClient) CreateExtensionVersion(version *ExtensionVersion) error {
	return client.Create(version).Error
}

func (client *DBClient) DeleteExtensionVersion(name, version string) error {
	return client.Where("name = ? and version =?", name, version).Delete(&ExtensionVersion{}).Error
}

func (client *DBClient) QueryExtensionVersions(name string, all string) ([]ExtensionVersion, error) {
	var result []ExtensionVersion
	query := client.Model(&ExtensionVersion{}).
		Where("name = ?", name)
	// 不显式指定all=true,只返回public的数据
	if all != "true" {
		query = query.Where("public = ?", true)
	}
	err := query.Find(&result).Error
	return result, err
}

func (client *DBClient) GetExtensionVersionCount(name string) (int64, error) {
	var count int64
	err := client.Model(&ExtensionVersion{}).
		Where("name = ? ", name).
		Count(&count).Error
	return count, err
}
