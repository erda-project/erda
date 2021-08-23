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
