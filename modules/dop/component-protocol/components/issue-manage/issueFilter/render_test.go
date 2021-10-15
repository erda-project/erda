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

package issueFilter

import (
	"testing"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/dop/services/issuefilterbm"
)

func Test_getMeta(t *testing.T) {
	data := map[string]interface{}{
		"meta": map[string]string{
			"id": "123",
		},
	}
	var m DeleteMeta
	assert.NoError(t, getMeta(data, &m))
	assert.Equal(t, "123", m.ID)
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

func Test_flushOptsByFilter(t *testing.T) {
	//ctx := context.WithValue(context.Background(), cptype.GlobalInnerKeyCtxSDK, &cptype.SDK{Tran: &MockTran{}})
	f := ComponentFilter{}
	// {"states":[657],"assigneeIDs":["2"]}
	assert.NoError(t, f.flushOptsByFilter("123", "eyJzdGF0ZXMiOls2NTddLCJhc3NpZ25lZUlEcyI6WyIyIl19"))
	assert.Equal(t, FrontendConditions{
		FilterID:    "123",
		States:      []int64{657},
		AssigneeIDs: []string{"2"},
	}, f.State.FrontendConditionValues)

	f.Bms = []issuefilterbm.MyFilterBm{
		{
			ID:           "123",
			FilterEntity: "eyJzdGF0ZXMiOls2NTddLCJhc3NpZ25lZUlEcyI6WyIyIl19",
		},
	}
	assert.NoError(t, f.flushOptsByFilterID("123"))
	assert.Equal(t, FrontendConditions{
		FilterID:    "123",
		States:      []int64{657},
		AssigneeIDs: []string{"2"},
	}, f.State.FrontendConditionValues)
}

func Test_determineFilterID(t *testing.T) {
	f := ComponentFilter{}
	assert.Equal(t, "", f.determineFilterID("eyJzdGF0ZXMiOls2NTddLCJhc3NpZ25lZUlEcyI6WyIyIl19"))

	f.Bms = []issuefilterbm.MyFilterBm{
		{
			ID:           "123",
			FilterEntity: "eyJzdGF0ZXMiOls2NTddLCJhc3NpZ25lZUlEcyI6WyIyIl19",
		},
	}
	assert.Equal(t, "123", f.determineFilterID("eyJzdGF0ZXMiOls2NTddLCJhc3NpZ25lZUlEcyI6WyIyIl19"))
}

func TestComponentFilter_setDefaultState(t *testing.T) {
	type args struct {
		stateMap map[string][]int64
		key      string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				stateMap: map[string][]int64{
					"t": {1},
				},
				key: "t",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &ComponentFilter{}
			f.setDefaultState(tt.args.stateMap, tt.args.key)
			assert.Equal(t, []int64{1}, f.State.FrontendConditionValues.States)
		})
		t.Run(tt.name, func(t *testing.T) {
			f := &ComponentFilter{
				State: State{
					FrontendConditionValues: FrontendConditions{
						States: []int64{2},
					},
				},
			}
			f.setDefaultState(tt.args.stateMap, tt.args.key)
			assert.Equal(t, []int64{2}, f.State.FrontendConditionValues.States)
		})
	}
}
