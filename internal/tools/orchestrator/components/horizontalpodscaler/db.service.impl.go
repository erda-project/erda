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

package horizontalpodscaler

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type dbServiceImpl struct {
	db *dbclient.DBClient
}

func (d *dbServiceImpl) CreateHPARule(runtimeHPA *dbclient.RuntimeHPA) error {
	return d.db.CreateRuntimeHPA(runtimeHPA)
}

func (d *dbServiceImpl) UpdateHPARule(runtimeHPA *dbclient.RuntimeHPA) error {
	return d.db.UpdateRuntimeHPA(runtimeHPA)
}

func (d *dbServiceImpl) GetRuntimeHPARulesByServices(id spec.RuntimeUniqueId, services []string) ([]dbclient.RuntimeHPA, error) {
	return d.db.GetRuntimeHPAByServices(id, services)
}

func (d *dbServiceImpl) DeleteRuntimeHPARulesByRuleId(ruleId string) error {
	if err := d.db.DeleteRuntimeHPAByRuleId(ruleId); err != nil {
		return err
	}
	return nil
}

func (d *dbServiceImpl) GetRuntimeHPARuleByRuleId(ruleId string) (dbclient.RuntimeHPA, error) {
	runtimeHPA, err := d.db.GetRuntimeHPARuleByRuleId(ruleId)
	if err != nil {
		return dbclient.RuntimeHPA{}, err
	}
	return *runtimeHPA, nil
}

func (d *dbServiceImpl) GetRuntimeHPARulesByRuntimeId(runtimeID uint64) ([]dbclient.RuntimeHPA, error) {
	runtimeHPAs, err := d.db.GetRuntimeHPARulesByRuntimeId(runtimeID)
	if err != nil {
		return []dbclient.RuntimeHPA{}, err
	}
	return runtimeHPAs, nil
}

func (d *dbServiceImpl) GetRuntime(id uint64) (*dbclient.Runtime, error) {
	return d.db.GetRuntime(id)
}

func (d *dbServiceImpl) GetPreDeployment(uniqueId spec.RuntimeUniqueId) (*dbclient.PreDeployment, error) {
	return d.db.FindPreDeployment(uniqueId)
}

func (d *dbServiceImpl) GetRuntimeHPAEventsByServices(runtimeId uint64, services []string) ([]dbclient.HPAEventInfo, error) {
	return d.db.GetRuntimeHPAEventsByServices(runtimeId, services)
}

func (d *dbServiceImpl) DeleteRuntimeHPAEventsByRuleId(ruleId string) error {
	if err := d.db.DeleteRuntimeHPAEventsByRuleId(ruleId); err != nil {
		return err
	}
	return nil
}

func newDBService(db *dbclient.DBClient) DBService {
	return &dbServiceImpl{db: db}
}

func NewDBService(orm *gorm.DB) DBService {
	return newDBService(&dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: orm,
		},
	})
}
