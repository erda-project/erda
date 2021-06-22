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
	"github.com/erda-project/erda/modules/cmdb/model"
)

func (client *DBClient) CreateAkSk(obj model.AkSk) (model.AkSk, error) {
	res := client.Create(&obj)
	if res.Error != nil {
		return model.AkSk{}, res.Error
	}
	return obj, nil
}

func (client *DBClient) ListAkSk(interval bool) ([]model.AkSk, error) {
	var objs []model.AkSk
	res := client.Where(&model.AkSk{IsInternal: interval}).Find(&objs)
	if res.Error != nil {
		return nil, res.Error
	}
	return objs, nil
}

func (client *DBClient) GetAkSkByAk(ak string) (model.AkSk, error) {
	var obj model.AkSk
	res := client.Where(&model.AkSk{Ak: ak}).First(&obj)
	if res.Error != nil {
		return model.AkSk{}, res.Error
	}
	return obj, nil
}

func (client *DBClient) DeleteAkSkByAk(ak string) error {
	return client.Where(&model.AkSk{Ak: ak}).Delete(&model.AkSk{}).Error
}
