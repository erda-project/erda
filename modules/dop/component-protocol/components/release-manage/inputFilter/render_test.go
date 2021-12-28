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

package inputFilter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/util"
)

func TestComponentInputFilter_GenComponentState(t *testing.T) {
	component := &cptype.Component{
		State: map[string]interface{}{
			"conditions": []Condition{
				{
					EmptyText:   "testText",
					Fixed:       true,
					Key:         "testKey",
					Label:       "testLabel",
					Placeholder: "testPlaceholder",
					Type:        "testType",
				},
			},
			"values": Values{
				Version: "testVersion",
			},
			"inputFilter__urlQuery": "testURLQuery",
		},
	}
	f := &ComponentInputFilter{}
	if err := f.GenComponentState(component); err != nil {
		t.Fatal(err)
	}
	isEqual, err := util.IsDeepEqual(f.State, component.State)
	if err != nil {
		t.Fatal(err)
	}
	if !isEqual {
		t.Errorf("test failed, state is not equal after generate")
		fmt.Println(component.State)
		fmt.Println(f.State)
	}
}

func getPair() (Values, string) {
	v := Values{Version: "testVersion"}
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	encode := base64.StdEncoding.EncodeToString(data)
	return v, encode
}

func TestComponentInputFilter_DecodeURLQuery(t *testing.T) {
	values, encode := getPair()
	c := ComponentInputFilter{
		sdk: &cptype.SDK{
			InParams: map[string]interface{}{
				"inputFilter__urlQuery": encode,
			},
		},
	}
	if err := c.DecodeURLQuery(); err != nil {
		t.Fatal(err)
	}
	if values.Version != c.State.Values.Version {
		t.Errorf("test failed, version is not expected after decode")
	}
}

func TestComponentInputFilter_EncodeURLQuery(t *testing.T) {
	values, encode := getPair()
	c := ComponentInputFilter{
		State: State{
			Values: values,
		},
	}
	if err := c.EncodeURLQuery(); err != nil {
		t.Fatal(err)
	}
	if encode != c.State.InputFilterURLQuery {
		t.Errorf("test failed, url query is not expected after encode")
	}
}

func TestComponentInputFilter_Transfer(t *testing.T) {
	c := ComponentInputFilter{
		State: State{
			Conditions: []Condition{
				{
					EmptyText:   "testText",
					Fixed:       true,
					Key:         "testKey",
					Label:       "testLabel",
					Placeholder: "testPlaceholder",
					Type:        "testType",
				},
			},
			Values: Values{
				Version: "testVersion",
			},
			InputFilterURLQuery: "testURLQuery",
		},
		Operations: map[string]interface{}{
			"testOp": Operation{
				Key:    "testKey",
				Reload: true,
			},
		},
	}
	component := &cptype.Component{}
	c.Transfer(component)
	isEqual, err := util.IsDeepEqual(c, component)
	if err != nil {
		t.Fatal(err)
	}
	if !isEqual {
		t.Errorf("test failed, component is not expected after transfer")
	}
}
