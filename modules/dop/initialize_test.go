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

package dop

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/endpoints"
)

func TestUpdateCmsNsConfigs(t *testing.T) {
	ep := endpoints.New()
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "PageListPipelineCrons",
		func(*bundle.Bundle, apistructs.PipelineCronPagingRequest) (*apistructs.PipelineCronPagingResponseData, error) {
			return &apistructs.PipelineCronPagingResponseData{
				Total: 3,
				Data: []*apistructs.PipelineCronDTO{
					{
						UserID: "1",
						OrgID:  1,
					},
					{
						UserID: "1",
						OrgID:  0,
					},
					{
						UserID: "",
						OrgID:  1,
					},
				},
			}, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "UpdatePipelineCron",
		func(*bundle.Bundle, apistructs.PipelineCronUpdateRequest) error {
			return nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(ep), "UpdateCmsNsConfigs",
		func(*endpoints.Endpoints, string, uint64) error {
			return nil
		})
	err := updateCmsNsConfigs(ep)
	assert.NoError(t, err)
}
