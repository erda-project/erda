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

func (client *DBClient) ListAccessKey(req apistructs.AccessKeyListQueryRequest) ([]model.AccessKey, error) {
	var objs []model.AccessKey
	query := client.Where(&model.AccessKey{
		Status:      req.Status,
		Subject:     req.Subject,
		SubjectType: req.SubjectType,
	})
	if req.IsSystem != nil {
		query = query.Where(map[string]interface{}{
			"is_system": req.IsSystem,
		})
	}
	res := query.Find(&objs)
	if res.Error != nil {
		return nil, res.Error
	}
	return objs, nil
}

func (client *DBClient) DeleteByAccessKeyID(ak string) error {
	return client.Where(&model.AccessKey{AccessKeyID: ak}).Delete(&model.AccessKey{}).Error
}
