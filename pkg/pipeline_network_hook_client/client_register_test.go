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
