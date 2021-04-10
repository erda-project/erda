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

package dbclient

import (
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/cimysql"
)

// APITestEnv 存储API测试环境变量
type APITestEnv struct {
	ID        int64     `xorm:"pk autoincr" json:"id"`
	CreatedAt time.Time `xorm:"created" json:"createdAt"`
	UpdatedAt time.Time `xorm:"updated" json:"updatedAt"`

	EnvID   int64  `xorm:"env_id" json:"envID"`
	EnvType string `xorm:"env_type" json:"envType"`
	Name    string `xorm:"name" json:"name"`
	Domain  string `xorm:"domain" json:"domain"`
	Header  string `xorm:"header" json:"header"`
	Global  string `xorm:"global" json:"global"`
}

// GetTestEnv 根据envID获取测试环境变量信息
func GetTestEnv(envID int64) (*APITestEnv, error) {
	testEnv := new(APITestEnv)
	_, err := cimysql.Engine.ID(envID).Get(testEnv)
	if err != nil {
		return nil, errors.Errorf("failed to get api test env info, ID:%d, (%+v)", envID, err)
	}

	return testEnv, nil
}

// GetTestEnvListByEnvID 根据envID获取测试环境变量信息
func GetTestEnvListByEnvID(envID int64, envType string) ([]APITestEnv, error) {
	testEnvList := []APITestEnv{}
	err := cimysql.Engine.Where("env_id = ? AND env_type = ?", envID, envType).Find(&testEnvList)
	if err != nil {
		return nil, errors.Errorf("failed to get api test env list, envID:%d, (%+v)", envID, err)
	}

	return testEnvList, nil
}
