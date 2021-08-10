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

package dto

import (
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
)

type GetOpenConsumersDto struct {
	DiceArgsDto
}

func (impl GetOpenConsumersDto) GenSelectOptions() []orm.SelectOption {
	var result []orm.SelectOption
	needCluster := true
	if impl.OrgId != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "org_id",
			Value:  impl.OrgId,
		})
	} else {
		needCluster = false
	}
	if impl.ProjectId != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "project_id",
			Value:  impl.ProjectId,
		})
	} else {
		needCluster = false
	}
	if impl.Env != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "env",
			Value:  impl.Env,
		})
	} else {
		needCluster = false
	}
	if needCluster {
		azDb, err := db.NewGatewayAzInfoServiceImpl()
		if err != nil {
			log.Error("create az db failed")
			return nil
		}
		az, err := azDb.GetAz(&orm.GatewayAzInfo{
			Env:       impl.Env,
			OrgId:     impl.OrgId,
			ProjectId: impl.ProjectId,
		})
		if err != nil {
			log.Error("get az failed")
			return nil
		}
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "az",
			Value:  az,
		})
	}
	return result
}
