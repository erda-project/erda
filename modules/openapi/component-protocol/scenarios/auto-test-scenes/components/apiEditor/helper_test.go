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

package apiEditor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

const testStepsData string = `[{"id":162,"type":"API","method":"","value":"{\"apiSpec\":{\"asserts\":[],\"body\":{\"content\":null,\"type\":\"\"},\"headers\":null,\"id\":\"\",\"method\":\"DELETE\",\"name\":\"deleteOrder\",\"out_params\":[{\"expression\":\"data.id\",\"key\":\"as\",\"source\":\"body:json\"}],\"params\":null,\"url\":\"/v2/store/order/{orderId}\"}}","name":"deleteOrder","preID":0,"preType":"Serial","sceneID":54,"spaceID":6,"creatorID":"","updaterID":"","Children":null,"apiSpecID":6},{"id":163,"type":"API","method":"","value":"{\"apiSpec\":{\"asserts\":[],\"body\":{\"content\":null,\"type\":\"\"},\"headers\":null,\"id\":\"\",\"method\":\"DELETE\",\"name\":\"deleteOrder\",\"out_params\":[{\"expression\":\"data.id\",\"key\":\"asd\",\"source\":\"body:json\"},{\"expression\":\"data.status\",\"key\":\"asd\",\"source\":\"body:json\"}],\"params\":[],\"url\":\"/sadfs/sad\"}}","name":"deleteOrder","preID":162,"preType":"Serial","sceneID":54,"spaceID":6,"creatorID":"","updaterID":"","Children":null,"apiSpecID":0}]`

func TestGetStepOutPut(t *testing.T) {
	var (
		err    error
		steps  []apistructs.AutoTestSceneStep
		output map[string]map[string]string
	)
	err = json.Unmarshal([]byte(testStepsData), &steps)
	output, err = GetStepOutPut(steps)

	assert.NoError(t, err)
	assert.Equal(t, "${{ outputs.162.as }}", output["#162-deleteOrder"]["as"])
	assert.Equal(t, "${{ outputs.163.asd }}", output["#163-deleteOrder"]["asd"])
}

func TestGenEmptyAPISpecStr(t *testing.T) {
	testEmptyAPISpec, testEmptyAPISpecStr := genEmptyAPISpecStr()
	assert.Equal(t, "GET", testEmptyAPISpec.APIInfo.Method)
	assert.Equal(t, `{"apiSpec":{"id":"","name":"","url":"","method":"GET","headers":null,"params":null,"body":{"type":"","content":null},"out_params":null,"asserts":null},"loop":null}`,
		testEmptyAPISpecStr)
}
