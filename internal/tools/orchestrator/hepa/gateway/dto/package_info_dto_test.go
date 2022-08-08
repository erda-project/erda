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

package dto_test

import (
	"sort"
	"testing"
	"time"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

func TestSortBySceneList_Less(t *testing.T) {
	var list = dto.SortBySceneList{
		{
			Id:       "",
			CreateAt: time.Now().Format(time.RFC3339),
			PackageDto: dto.PackageDto{
				Name:  "hub-0",
				Scene: orm.HubScene,
			},
		}, {
			Id:       "",
			CreateAt: time.Now().Format(time.RFC3339),
			PackageDto: dto.PackageDto{
				Name:  "hub-1",
				Scene: orm.HubScene,
			},
		}, {
			Id:       "",
			CreateAt: time.Now().Format(time.RFC3339),
			PackageDto: dto.PackageDto{
				Name:  "unity",
				Scene: orm.UnityScene,
			},
		}, {
			Id:       "",
			CreateAt: time.Now().Format(time.RFC3339),
			PackageDto: dto.PackageDto{
				Name:  "xxx",
				Scene: orm.WebapiScene,
			},
		}, {
			Id:       "",
			CreateAt: time.Now().Format(time.RFC3339),
			PackageDto: dto.PackageDto{
				Name:  "yyy",
				Scene: orm.WebapiScene,
			},
		}, {
			Id:       "",
			CreateAt: time.Now().Format(time.RFC3339),
			PackageDto: dto.PackageDto{
				Name:  "zzz",
				Scene: orm.OpenapiScene,
			},
		}, {
			Id:       "",
			CreateAt: time.Now().Format(time.RFC3339),
			PackageDto: dto.PackageDto{
				Name:  "aaa",
				Scene: orm.OpenapiScene,
			},
		},
	}

	sort.Sort(list)
	t.Log(list)
}
