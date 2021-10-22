//  Copyright (c) 2021 Terminus, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (e *Endpoints) ResourceOverviewReport(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var data = apistructs.ResourceOverviewReportData{
		Total: 5,
		List: []*apistructs.ResourceOverviewReportDataItem{
			{
				ProjectID:          1,
				ProjectName:        "project-1",
				ProjectDisplayName: "项目-1",
				OwnerUserID:        2,
				OwnerUserName:      "dice",
				OwnerUserNickName:  "dice",
				CPUQuota:           100.0,
				CPUWaterLevel:      0.64,
				MemQuota:           100,
				MemWaterLevel:      0.64,
				Nodes:              10,
			}, {
				ProjectID:          2,
				ProjectName:        "project-2",
				ProjectDisplayName: "项目-2",
				OwnerUserID:        2,
				OwnerUserName:      "dice",
				OwnerUserNickName:  "dice",
				CPUQuota:           200.00,
				CPUWaterLevel:      0.45,
				MemQuota:           156,
				MemWaterLevel:      0.68,
				Nodes:              9,
			}, {
				ProjectID:          2,
				ProjectName:        "project-3",
				ProjectDisplayName: "项目-3",
				OwnerUserID:        2,
				OwnerUserName:      "dice",
				OwnerUserNickName:  "dice",
				CPUQuota:           45,
				CPUWaterLevel:      0.86,
				MemQuota:           47,
				MemWaterLevel:      0.46,
				Nodes:              3,
			}, {
				ProjectID:          2,
				ProjectName:        "project-4",
				ProjectDisplayName: "项目-4",
				OwnerUserID:        2,
				OwnerUserName:      "dice",
				OwnerUserNickName:  "dice",
				CPUQuota:           45,
				CPUWaterLevel:      0.86,
				MemQuota:           47,
				MemWaterLevel:      0.46,
				Nodes:              3,
			},
			{
				ProjectID:          2,
				ProjectName:        "project-5",
				ProjectDisplayName: "项目-5",
				OwnerUserID:        2,
				OwnerUserName:      "dice",
				OwnerUserNickName:  "dice",
				CPUQuota:           45,
				CPUWaterLevel:      0.86,
				MemQuota:           47,
				MemWaterLevel:      0.46,
				Nodes:              3,
			},
		},
	}

	// todo: authentication

	return httpserver.OkResp(data)
}
