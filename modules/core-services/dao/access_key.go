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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
)

func (client *DBClient) CreateAccessKey(obj model.AccessKey) (model.AccessKey, error) {
	res := client.Create(&obj)
	if res.Error != nil {
		return model.AccessKey{}, res.Error
	}
	return obj, nil
}

func (client *DBClient) UpdateAccessKey(ak string, req apistructs.AccessKeyUpdateRequest) (model.AccessKey, error) {
	res := client.Model(&model.AccessKey{}).Where("access_key_id = ?", ak).Update(model.AccessKey{
		Status:      req.Status,
		Description: req.Description,
	})
	if res.Error != nil {
		return model.AccessKey{}, res.Error
	}
	return model.AccessKey{}, nil
}

func (client *DBClient) ListSystemAccessKey(isSystem bool) ([]model.AccessKey, error) {
	var objs []model.AccessKey
	res := client.Where(&model.AccessKey{IsSystem: isSystem, Status: apistructs.AccessKeyStatusActive}).Find(&objs)
	if res.Error != nil {
		return nil, res.Error
	}
	return objs, nil
}

func (client *DBClient) GetByAccessKeyID(ak string) (model.AccessKey, error) {
	var obj model.AccessKey
	res := client.Where(&model.AccessKey{AccessKeyID: ak}).First(&obj)
	if res.Error != nil {
		return model.AccessKey{}, res.Error
	}
	return obj, nil
}

func (client *DBClient) DeleteByAccessKeyID(ak string) error {
	return client.Where(&model.AccessKey{AccessKeyID: ak}).Delete(&model.AccessKey{}).Error
}
