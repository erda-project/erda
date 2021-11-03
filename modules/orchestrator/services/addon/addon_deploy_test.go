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

package addon

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestMysqlPreProcess(t *testing.T) {
	params := &apistructs.AddonHandlerCreateItem{Plan: apistructs.AddonBasic}
	addonSpec := &apistructs.AddonExtension{Plan: map[string]apistructs.AddonPlanItem{
		apistructs.AddonBasic: {
			Nodes: 2,
		},
	}}
	addonDeployGroup := &apistructs.ServiceGroupCreateV2Request{
		GroupLabels: map[string]string{
			"ADDON_GROUPS": "2",
		},
	}
	mysqlPreProcess(params, addonSpec, addonDeployGroup)
	assert.Equal(t, "1", addonDeployGroup.GroupLabels["ADDON_GROUPS"])
	assert.Equal(t, 1, addonSpec.Plan[apistructs.AddonBasic].Nodes)
}
