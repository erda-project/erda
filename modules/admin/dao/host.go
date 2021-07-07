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
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/cmdb/model"
)

// GetHostByClusterAndIP get host info according cluster & privateAddr
func (client *DBClient) GetHostByClusterAndIP(clusterName, privateAddr string) (*model.Host, error) {
	var host model.Host
	if err := client.Where("cluster = ?", clusterName).
		Where("private_addr = ?", privateAddr).First(&host).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return &host, nil
}
