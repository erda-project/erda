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

package assetsvc

import (
	"github.com/erda-project/erda/apistructs"
)

func onceADayLimitType() []apistructs.LimitType {
	once := 1
	return []apistructs.LimitType{{
		Day:    &once,
		Hour:   nil,
		Minute: nil,
		Second: nil,
	}}
}

func unlimitedSLA(access *apistructs.APIAccessesModel) *apistructs.SLAModel {
	return &apistructs.SLAModel{
		BaseModel: apistructs.BaseModel{
			ID:        0,
			CreatedAt: access.CreatedAt,
			UpdatedAt: access.CreatedAt,
			CreatorID: access.CreatorID,
			UpdaterID: access.CreatorID,
		},
		Name:     apistructs.UnlimitedSLAName,
		Desc:     "不限制流量 SLA",
		Approval: apistructs.AuthorizationManual,
		AccessID: access.ID,
		Source:   apistructs.SourceSystem,
	}
}

// 如果传入的 slaID 为 0, 返回无限制流量的 SLA; 否则查询数据库返回 slaID 对应的库表记录
func (svc *Service) querySLAByID(slaID uint64, access *apistructs.APIAccessesModel) (*apistructs.SLAModel, error) {
	if slaID == 0 {
		return unlimitedSLA(access), nil
	}
	var sla apistructs.SLAModel
	if err := svc.FirstRecord(&sla, map[string]interface{}{"id": slaID}); err != nil {
		return nil, err
	}
	return &sla, nil
}
