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

package filter

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func getTestURLQuery() (Values, string) {
	v := Values{
		Kind:      []string{"test"},
		Namespace: []string{"test"},
		Status:    []string{"test"},
		Search:    "test",
	}
	data, _ := json.Marshal(v)
	encode := base64.StdEncoding.EncodeToString(data)
	return v, encode
}

func isEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestComponentFilter_DecodeURLQuery(t *testing.T) {
	values, res := getTestURLQuery()
	filter := &ComponentFilter{
		ctxBdl: protocol.ContextBundle{
			InParams: map[string]interface{}{
				"filter__urlQuery": res,
			},
		},
	}
	if err := filter.DecodeURLQuery(); err != nil {
		t.Errorf("test failed, %v", err)
	}
	if !isEqual(filter.State.Values.Namespace, values.Namespace) || !isEqual(filter.State.Values.Status, values.Status) ||
		!isEqual(filter.State.Values.Kind, values.Kind) || filter.State.Values.Search != values.Search {
		t.Errorf("test failed, edcode result is not expected")
	}
}

func TestComponentFilter_GenComponentState(t *testing.T) {
	component := &apistructs.Component{
		State: map[string]interface{}{
			"clusterName": "test",
			"conditions": []Condition{
				{
					Key:         "test",
					Label:       "test",
					Placeholder: "test",
					Type:        "test",
					Fixed:       true,
					Options: []Option{
						{
							Label: "test",
							Value: "test",
						},
					},
				},
			},
			"values": Values{
				Kind:      []string{"test"},
				Namespace: []string{"test"},
				Status:    []string{"test"},
				Search:    "test",
			},
		},
	}
	src, err := json.Marshal(component.State)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}

	f := &ComponentFilter{}
	if err := f.GenComponentState(component); err != nil {
		t.Errorf("test failed, %v", err)
	}

	dst, err := json.Marshal(f.State)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}

	if string(src) != string(dst) {
		t.Error("test failed, generate result is unexpected")
	}
}

func TestComponentFilter_EncodeURLQuery(t *testing.T) {
	values, data := getTestURLQuery()
	f := ComponentFilter{
		State: State{
			Values: values,
		},
	}
	if err := f.EncodeURLQuery(); err != nil {
		t.Errorf("test failed, %v", err)
	}
	if f.State.FilterURLQuery != data {
		t.Error("test failed, encode url query result is unexpected")
	}
}

func TestHasSuffix(t *testing.T) {
	dev := "project-1-dev"
	_, ok := hasSuffix(dev)
	if !ok {
		t.Error("test failed, expected to have suffix \"-dev\", actual not")
	}

	stage := "project-2-staging"
	_, ok = hasSuffix(stage)
	if !ok {
		t.Error("test failed, expected to have suffix \"-staging\", actual not")
	}

	test := "project-3-test"
	_, ok = hasSuffix(test)
	if !ok {
		t.Error("test failed, expected to have suffix \"-test\", actual not")
	}

	prod := "project-4-prod"
	_, ok = hasSuffix(prod)
	if !ok {
		t.Error("test failed, expected to have suffix \"-prod\", actual not")
	}

	other := "default"
	_, ok = hasSuffix(other)
	if ok {
		t.Error("test failed, expected to not have suffix, actual do")
	}
}
