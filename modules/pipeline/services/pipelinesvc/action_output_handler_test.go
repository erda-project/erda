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

package pipelinesvc

import (
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/encoding/jsonpath"
)

func Test_handlerActionOutputsWithJq(t *testing.T) {
	var table = []struct {
		action  *apistructs.PipelineYmlAction
		outputs map[string]interface{}
	}{
		{
			action: &apistructs.PipelineYmlAction{},
			outputs: map[string]interface{}{
				"name":  "",
				"name1": "",
				"key1":  "",
			},
		},
		{
			action:  &apistructs.PipelineYmlAction{},
			outputs: map[string]interface{}{},
		},
		{
			action:  &apistructs.PipelineYmlAction{},
			outputs: nil,
		},
	}

	for _, data := range table {

		monkey.Patch(jsonpath.JQ, func(jsonInput, filter string) (interface{}, error) {
			var outputs []string
			for key, _ := range data.outputs {
				outputs = append(outputs, key)
			}
			return outputs, nil
		})

		data.action.Params = data.outputs
		outputs, err := handlerActionOutputsWithJq(data.action, "test")
		assert.NoError(t, err)

		assert.Equal(t, len(outputs), len(data.outputs))
		for _, output := range outputs {
			_, ok := data.outputs[output]
			assert.True(t, ok)
		}
	}
}
