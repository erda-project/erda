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

// ApiTest 存储pmp上接口测试相关的信息，对应数据库表dice_api_test
type ApiTest struct {
	ID        int64     `xorm:"pk autoincr" json:"id"`
	CreatedAt time.Time `xorm:"created" json:"createdAt"`
	UpdatedAt time.Time `xorm:"updated" json:"updatedAt"`

	UsecaseID    int64  `xorm:"usecase_id" json:"usecaseID"`
	UsecaseOrder int64  `xorm:"usecase_order" json:"usecaseOrder"`
	ProjectID    int64  `xorm:"project_id" json:"projectID"`
	PipelineID   int64  `xorm:"pipeline_id" json:"pipelineID"`
	Status       string `xorm:"status" json:"status"`
	ApiInfo      string `xorm:"text" json:"apiInfo"`
	ApiRequest   string `xorm:"longtext" json:"apiRequest"`
	ApiResponse  string `xorm:"longtext" json:"apiResponse"`
	AssertResult string `xorm:"text" json:"assertResult"`
}

// TableName DiceApiTest对应的数据库表dice_api_test
func (ApiTest) TableName() string {
	return "dice_api_test"
}

// CreateApiTest 创建apiTest信息
func CreateApiTest(at *ApiTest) (int64, error) {
	ID, err := cimysql.Engine.InsertOne(at)
	if err != nil {
		return 0, errors.Errorf("failed to insert api test info, (%+v)", err)
	}

	return ID, nil
}

// UpdateApiTest 更新apiTest信息
func UpdateApiTest(at *ApiTest) (int64, error) {
	ID, err := cimysql.Engine.Id(at.ID).Update(at)
	if err != nil {
		return 0, errors.Errorf("failed to update api test info, ID:%d, (%+v)", at.ID, err)
	}

	return ID, nil
}

// UpdateApiTestResults 更新apiTest结果信息：status，api_response，assert_result
func UpdateApiTestResults(at *ApiTest) (int64, error) {
	ID, err := cimysql.Engine.Id(at.ID).Cols("status", "api_request", "api_response", "assert_result").Update(at)
	if err != nil {
		return 0, errors.Errorf("failed to update api test info, ID:%d, (%+v)", at.ID, err)
	}

	return ID, nil
}

// GetApiTest 根据apiID获取apiTest信息
func GetApiTest(apiID int64) (*ApiTest, error) {
	apiTest := new(ApiTest)
	_, err := cimysql.Engine.ID(apiID).Get(apiTest)
	if err != nil {
		return nil, errors.Errorf("failed to get api test info, ID:%d, (%+v)", apiID, err)
	}

	return apiTest, nil
}

// DeleteApiTest 删除apiTest信息
func DeleteApiTest(apiID int64) error {
	apiTest := new(ApiTest)
	_, err := cimysql.Engine.ID(apiID).Delete(apiTest)
	if err != nil {
		return errors.Errorf("failed to get api test info, ID:%d, (%+v)", apiID, err)
	}

	return nil
}

// GetApiTestListByUsecaseID 根据usecaseID获取apiTest列表
func GetApiTestListByUsecaseID(usecaseID int64) ([]ApiTest, error) {
	apiTestsList := []ApiTest{}
	err := cimysql.Engine.Where("usecase_id = ?", usecaseID).Asc("usecase_order").Find(&apiTestsList)
	if err != nil {
		return nil, errors.Errorf("failed to get api test list, usecaseID:%d, (%+v)", usecaseID, err)
	}

	return apiTestsList, nil
}
