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

package pipeline_network_hook_client

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"gotest.tools/assert"

	"github.com/xormplus/xorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
)

func TestRegisterLifecycleHookClient(t *testing.T) {
	var table = []struct {
		clients []*dbclient.PipelineLifecycleHookClient
	}{
		{
			clients: []*dbclient.PipelineLifecycleHookClient{
				{
					Name:   "FDP",
					ID:     1,
					Host:   "FDP.svc.default",
					Prefix: "/api/workflow",
				},
				{
					Name:   "dice",
					ID:     2,
					Host:   "dice.svc.default",
					Prefix: "/api/dice",
				},
			},
		},
		{
			clients: []*dbclient.PipelineLifecycleHookClient{},
		},
	}

	for _, v := range table {
		var e dbclient.Client
		var engine xorm.Engine
		guard1 := monkey.PatchInstanceMethod(reflect.TypeOf(&engine), "Find", func(engine *xorm.Engine, beans interface{}, condiBeans ...interface{}) error {
			checkRunResultResponseJson, _ := json.Marshal(v.clients)
			buffer := bytes.NewBuffer(checkRunResultResponseJson)
			err := json.NewDecoder(buffer).Decode(&beans)
			return err
		})
		e.Engine = &engine
		defer guard1.Unpatch()
		RegisterLifecycleHookClient(&e)

		assert.Equal(t, len(v.clients), len(hookClientMap))
		for _, client := range v.clients {
			assert.Equal(t, hookClientMap[client.Name].Name, client.Name)
			assert.Equal(t, hookClientMap[client.Name].ID, client.ID)
			assert.Equal(t, hookClientMap[client.Name].Prefix, client.Prefix)
			assert.Equal(t, hookClientMap[client.Name].Host, client.Host)
		}
		hookClientMap = map[string]*apistructs.PipelineLifecycleHookClient{}
	}
}
