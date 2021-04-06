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

// Package runtime 应用实例相关操作
package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestModifyStatusIfNotForDisplay(t *testing.T) {
	runtime := apistructs.RuntimeInspectDTO{
		Status: "Unknown",
		Services: map[string]*apistructs.RuntimeInspectServiceDTO{
			"test": {
				Status: "Stopped",
			},
		},
	}
	updateStatusToDisplay(&runtime)
	assert.Equal(t, "Unknown", runtime.Status)
	for _, s := range runtime.Services {
		assert.Equal(t, "Stopped", s.Status)
	}
}
