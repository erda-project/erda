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

package executeTaskTable

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestComponentFilter_ImportExport(t *testing.T) {
	res := apistructs.AutoTestSceneStep{}
	var b []byte = []byte("{\"id\":13,\"type\":\"SCENE\",\"method\":\"\",\"value\":\"{\\\"runParams\\\":{},\\\"sceneID\\\":45}\",\"name\":\"嵌套：Dice 官网\",\"preID\":0,\"preType\":\"Serial\",\"sceneID\":45,\"spaceID\":42,\"creatorID\":\"\",\"updaterID\":\"\",\"Children\":null}")
	err := json.Unmarshal(b, &res)
	t.Log(res)
	assert.NoError(t, err)

}
