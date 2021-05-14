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

package pipelineyml

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
)

func TestVisitorParamsBigInt(t *testing.T) {
	s := `
version: 1.1
params:
  - name: a
    required: true
    default: 13245555555555
    type: int
stages:
  - stage:
      - api-test:
          version: "2.0"
          params:
            body:
              type: none
            method: GET
            params:
              - key: a
                value: ${params.a}
            url: test?a=${params.a}
`
	y, err := New([]byte(s),
		//WithRunParams([]apistructs.PipelineRunParam{
		//	{
		//		Name:  "a",
		//		Value: 13245555555555,
		//	},
		//}),
		WithFlatParams(true),
	)
	assert.NoError(t, err)

	spew.Dump(y.Spec().Stages[0])
	fmt.Println(y.Spec().Stages[0].Actions[0]["api-test"].Params["url"])
}

func TestXXX(t *testing.T) {
	runParams := []apistructs.PipelineRunParam{
		{
			Name:  "a",
			Value: 13245555555555,
		},
	}
	b, err := yaml.Marshal(runParams)
	assert.NoError(t, err)
	err = yaml.Unmarshal(b, &runParams)
	assert.NoError(t, err)
	spew.Dump(runParams)
}

func TestFloat64(t *testing.T) {
	s := struct {
		F interface{} `json:"f,omitempty"`
	}{
		F: 132455555555555,
	}
	spew.Dump(s)
	b, err := json.Marshal(s)
	assert.NoError(t, err)
	err = json.Unmarshal(b, &s)
	assert.NoError(t, err)
	spew.Dump(s)
	fmt.Println(fmt.Sprintf("%v", s.F))

	var replaceStr string
	switch v := s.F.(type) {
	case int:
		replaceStr = strconv.Itoa(v)
	case float64:
		if float64(int64(v)) == v {
			replaceStr = fmt.Sprintf("%.f", s.F)
		} else {
			replaceStr = fmt.Sprintf("%.2f", s.F)
		}
	case float32:
		replaceStr = fmt.Sprintf("%v", v)
	case bool:
		replaceStr = strconv.FormatBool(v)
	case string:
		replaceStr = v
	}
	fmt.Println(replaceStr)

	fmt.Println(strconv.FormatFloat(1324555555555, 'f', -1, 64))
	fmt.Println(strconv.FormatFloat(1, 'f', -1, 64))
	fmt.Println(strconv.FormatFloat(1.1, 'f', -1, 64))
	fmt.Println(strconv.FormatFloat(1.0, 'f', -1, 64))
	fmt.Println(strconv.FormatFloat(11.1111, 'f', -1, 64))
}
