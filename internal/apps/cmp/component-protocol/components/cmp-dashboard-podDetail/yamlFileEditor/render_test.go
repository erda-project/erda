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

package yamlFileEditor

import (
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/apps/cmp/component-protocol/cputil"
)

func TestComponentYamlFileEditor_GenComponentState(t *testing.T) {
	c := &cptype.Component{State: map[string]interface{}{
		"clusterName": "testClusterName",
		"podId":       "testPodID",
		"value":       "testValue",
	}}
	component := &ComponentYamlFileEditor{}
	if err := component.GenComponentState(c); err != nil {
		t.Fatal(err)
	}

	ok, err := cputil.IsDeepEqual(c.State, component.State)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("test failed, json is not equal")
	}
}

type MockTran struct {
	i18n.Translator
}

func (m *MockTran) Text(lang i18n.LanguageCodes, key string) string {
	return ""
}

func (m *MockTran) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return ""
}

func TestComponentYamlFileEditor_SetComponentValue(t *testing.T) {
	component := &ComponentYamlFileEditor{
		sdk: &cptype.SDK{
			Tran: &MockTran{},
		},
	}
	component.SetComponentValue()
	if _, ok := component.Operations["submit"]; !ok || !component.Props.Bordered ||
		len(component.Props.FileValidate) != 2 {
		t.Errorf("test failed")
	}
}

func TestComponentYamlFileEditor_Transfer(t *testing.T) {
	component := &ComponentYamlFileEditor{
		Props: Props{
			Bordered:     true,
			FileValidate: []string{"test"},
			MinLines:     22,
			Actions: Actions{
				Copy: true,
			},
		},
		State: State{
			ClusterName: "testClusterName",
			PodID:       "testPodID",
			Value:       "testValue",
		},
		Operations: map[string]interface{}{
			"testOP": Operation{
				Key:    "testKey",
				Reload: true,
			},
		},
	}

	c := &cptype.Component{}
	component.Transfer(c)
	ok, err := cputil.IsDeepEqual(c, component)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("test failed, json is not equal")
	}
}
