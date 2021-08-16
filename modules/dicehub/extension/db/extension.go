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
	"errors"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/gorm"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
)

// ExtensionConfig .
type ExtensionConfigDB struct {
	*gorm.DB
}

func (ext *ExtensionVersion) ToApiData(typ string, yamlFormat bool) (*pb.ExtensionVersion, error) {
	if yamlFormat {
		return &pb.ExtensionVersion{
			Name:      ext.Name,
			Type:      typ,
			Version:   ext.Version,
			Dice:      structpb.NewStringValue(ext.Dice),
			Spec:      structpb.NewStringValue(ext.Spec),
			Swagger:   structpb.NewStringValue(ext.Swagger),
			Readme:    ext.Readme,
			CreatedAt: timestamppb.New(ext.CreatedAt),
			UpdatedAt: timestamppb.New(ext.UpdatedAt),
			IsDefault: ext.IsDefault,
			Public:    ext.Public,
		}, nil
	} else {
		diceData, err := yaml.YAMLToJSON([]byte(ext.Dice))
		if err != nil {
			return nil, err
		}
		specData, err := yaml.YAMLToJSON([]byte(ext.Spec))
		if err != nil {
			return nil, err
		}
		swaggerData, err := yaml.YAMLToJSON([]byte(ext.Swagger))
		if err != nil {
			return nil, err
		}
		dice := &structpb.Value{}
		err = dice.UnmarshalJSON(diceData)
		if err != nil {
			return nil, err
		}
		spec := &structpb.Value{}
		err = spec.UnmarshalJSON(specData)
		if err != nil {
			return nil, err
		}
		swag := &structpb.Value{}
		err = swag.UnmarshalJSON(swaggerData)
		if err != nil {
			return nil, err
		}
		return &pb.ExtensionVersion{
			Name:      ext.Name,
			Type:      typ,
			Version:   ext.Version,
			Dice:      dice,
			Spec:      spec,
			Swagger:   swag,
			Readme:    ext.Readme,
			CreatedAt: timestamppb.New(ext.CreatedAt),
			UpdatedAt: timestamppb.New(ext.UpdatedAt),
			IsDefault: ext.IsDefault,
			Public:    ext.Public,
		}, nil
	}
}

func (client *ExtensionConfigDB) CreateExtension(extension *Extension) error {
	var cnt int64
	client.Model(&Extension{}).Where("name = ?", extension.Name).Count(&cnt)
	if cnt == 0 {
		err := client.Create(extension).Error
		return err
	} else {
		return errors.New("name already exist")
	}
}

func (client *ExtensionConfigDB) QueryExtensions(all string, typ string, labels string) ([]Extension, error) {
	var result []Extension
	query := client.Model(&Extension{})

	// if all != true,only return data with public = true
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

func (client *ExtensionConfigDB) GetExtension(name string) (*Extension, error) {
	var result Extension
	err := client.Model(&Extension{}).Where("name = ?", name).Find(&result).Error
	return &result, err
}

func (client *ExtensionConfigDB) DeleteExtension(name string) error {
	return client.Where("name = ?", name).Delete(&Extension{}).Error
}

func (client *ExtensionConfigDB) GetExtensionVersion(name string, version string) (*ExtensionVersion, error) {
	var result ExtensionVersion
	err := client.Model(&ExtensionVersion{}).
		Where("name = ? ", name).
		Where("version = ?", version).
		Find(&result).Error
	return &result, err
}

func (client *ExtensionConfigDB) GetExtensionDefaultVersion(name string) (*ExtensionVersion, error) {
	var result ExtensionVersion
	err := client.Model(&ExtensionVersion{}).
		Where("name = ? ", name).
		Where("is_default = ? ", true).
		Limit(1).
		Find(&result).Error
	//no default,find latest update & public = true
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

func (client *ExtensionConfigDB) SetUnDefaultVersion(name string) error {
	return client.Model(&ExtensionVersion{}).
		Where("is_default = ?", true).
		Where("name = ?", name).
		Update("is_default", false).Error
}

func (client *ExtensionConfigDB) CreateExtensionVersion(version *ExtensionVersion) error {
	return client.Create(version).Error
}

func (client *ExtensionConfigDB) DeleteExtensionVersion(name, version string) error {
	return client.Where("name = ? and version =?", name, version).Delete(&ExtensionVersion{}).Error
}

func (client *ExtensionConfigDB) QueryExtensionVersions(name string, all string) ([]ExtensionVersion, error) {
	var result []ExtensionVersion
	query := client.Model(&ExtensionVersion{}).
		Where("name = ?", name)
	// if all != true,only return data with public = true
	if all != "true" {
		query = query.Where("public = ?", true)
	}
	err := query.Find(&result).Error
	return result, err
}

func (client *ExtensionConfigDB) GetExtensionVersionCount(name string) (int64, error) {
	var count int64
	err := client.Model(&ExtensionVersion{}).
		Where("name = ? ", name).
		Count(&count).Error
	return count, err
}

func (client *ExtensionConfigDB) QueryAllExtensions() ([]ExtensionVersion, error) {
	var result []ExtensionVersion
	err := client.Model(&ExtensionVersion{}).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}
