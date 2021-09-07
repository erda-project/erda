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

package query

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/extensions/loghub/index/query/db"
	db2 "github.com/erda-project/erda/modules/msp/instance/db"
)

func TestGetLogIndices_WithNoneEmptyOrgId_Should_Return_Indices_With_OrgAlias(t *testing.T) {
	result := getLogIndices("rlogs-", "1")
	if len(result) == 0 {
		t.Errorf("should not return empty slice")
	}

	if result[0] != "rlogs-org-1" {
		t.Errorf("should return org alias")
	}
}

//go:generate mockgen -destination=./clients_mock_logs_test.go -package query github.com/erda-project/erda-infra/base/logs Logger
func TestGetESClientsFromLogAnalyticsByCluster_Should_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := &provider{
		L:     NewMockLogger(ctrl),
		C:     &config{},
		mysql: &gorm.DB{},
		db: &db.DB{
			LogDeployment:        db.LogDeploymentDB{},
			LogServiceInstanceDB: db.LogServiceInstanceDB{},
			LogInstanceDB:        db.LogInstanceDB{},
		},
		bdl: bundle.New(),
	}

	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "QueryByOrgIDAndClusters")
	monkey.PatchInstanceMethod(reflect.TypeOf(&p.db.LogDeployment), "QueryByOrgIDAndClusters", func(_ *db.LogDeploymentDB, orgID int64, clusters ...string) ([]*db.LogDeployment, error) {
		return []*db.LogDeployment{
			{LogType: string(db2.LogTypeLogService), ESURL: "http://localhost:9200"},
		}, nil
	})

	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&p.db.LogInstanceDB), "GetByLogKey")
	monkey.PatchInstanceMethod(reflect.TypeOf(&p.db.LogInstanceDB), "GetByLogKey", func(_ *db.LogInstanceDB, logKey string) (*db.LogInstance, error) {
		return &db.LogInstance{LogType: string(db2.LogTypeLogAnalytics), LogKey: "logKey-1", Config: `{"MSP_ENV_ID":"msp_env_id_1"}`}, nil
	})

	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&p.db.LogInstanceDB), "GetListByClusterAndProjectIdAndWorkspace")
	monkey.PatchInstanceMethod(reflect.TypeOf(&p.db.LogInstanceDB), "GetListByClusterAndProjectIdAndWorkspace", func(_ *db.LogInstanceDB, clusterName, projectId, workspace string) ([]db.LogInstance, error) {
		return []db.LogInstance{
			{LogType: string(db2.LogTypeLogService), LogKey: "logKey-3", Config: `{"MSP_ENV_ID":"msp_env_id_1"}`},
			{LogType: string(db2.LogTypeLogService), LogKey: "logKey-2", Config: `{"MSP_ENV_ID":"msp_env_id_1"}`},
			{LogType: string(db2.LogTypeLogAnalytics), LogKey: "logKey-1", Config: `{"MSP_ENV_ID":"msp_env_id_1"}`},
		}, nil
	})

	p.getESClientsFromLogAnalyticsByCluster(1, "addon-1", "cluster-1")
}
