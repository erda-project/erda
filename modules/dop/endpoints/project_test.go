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

package endpoints

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestAddOnsFilterIn(t *testing.T) {
	addOns := []apistructs.AddonFetchResponseData{
		{
			ID:                  "1",
			PlatformServiceType: 1,
		},
		{
			ID:                  "2",
			PlatformServiceType: 0,
		},
		{
			ID:                  "3",
			PlatformServiceType: 1,
		},
	}
	newAddOns := addOnsFilterIn(addOns, func(addOn *apistructs.AddonFetchResponseData) bool {
		return addOn.PlatformServiceType == 0
	})
	assert.Equal(t, 1, len(newAddOns))
}
