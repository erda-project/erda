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

package instance

import (
	"encoding/json"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/modules/msp/instance/db"
)

// Instance .
type Instance struct {
	Instance       *db.InstanceDB
	InstanceTenant *db.InstanceTenantDB
}

// New .
func New(gdb *gorm.DB) *Instance {
	return &Instance{
		Instance:       &db.InstanceDB{DB: gdb},
		InstanceTenant: &db.InstanceTenantDB{DB: gdb},
	}
}

// Config .
type ConfigOption struct {
	Config  map[string]interface{}
	Options map[string]interface{}
}

// GetConfigGroup .
func (ins *Instance) GetConfigOptionByGroup(engine, group string) (*ConfigOption, error) {
	tenant, err := ins.InstanceTenant.GetByFields(map[string]interface{}{
		"Engine":      engine,
		"TenantGroup": group,
	})
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, nil
	}
	co := &ConfigOption{}
	if len(tenant.Config) > 0 {
		co.Config = make(map[string]interface{})
		json.Unmarshal([]byte(tenant.Config), &co.Config)
	}
	if len(tenant.Options) > 0 {
		co.Options = make(map[string]interface{})
		json.Unmarshal([]byte(tenant.Options), &co.Options)
	}
	instance, err := ins.Instance.GetByID(tenant.InstanceID)
	if err != nil {
		return nil, err
	}
	if instance != nil {
		if len(instance.Config) > 0 {
			config := make(map[string]interface{})
			json.Unmarshal([]byte(instance.Config), &config)
			co.Config = mergeMap(config, co.Config)
		}
		if len(instance.Options) > 0 {
			opts := make(map[string]interface{})
			json.Unmarshal([]byte(instance.Options), &opts)
			co.Options = mergeMap(opts, co.Options)
		}
	}
	return co, nil
}

func mergeMap(list ...map[string]interface{}) map[string]interface{} {
	if len(list) < 0 {
		return nil
	}
	merge := make(map[string]interface{})
	for _, item := range list {
		for k, v := range item {
			merge[k] = v
		}
	}
	return merge
}
