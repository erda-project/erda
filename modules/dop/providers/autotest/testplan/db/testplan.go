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

package db

import (
	"github.com/jinzhu/gorm"
)

// TestPlanDB .
type TestPlanDB struct {
	*gorm.DB
}

// UpdateTestPlanV2 Update test plan
func (client *TestPlanDB) UpdateTestPlanV2(testPlanID uint64, fields map[string]interface{}) error {
	tp := TestPlanV2{}
	tp.ID = testPlanID

	return client.Model(&tp).Updates(fields).Error
}

// CreateAutoTestExecHistory .
func (db *TestPlanDB) CreateAutoTestExecHistory(execHistory *AutoTestExecHistory) error {
	return db.Create(execHistory).Error
}

// BatchCreateAutoTestExecHistory .
func (db *TestPlanDB) BatchCreateAutoTestExecHistory(list []AutoTestExecHistory) error {
	return db.Create(list).Error
}
