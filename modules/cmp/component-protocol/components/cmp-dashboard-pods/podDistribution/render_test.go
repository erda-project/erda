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

package PodDistribution

import (
	"context"
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
)

type MockTran struct {
	i18n.Translator
}

func (m *MockTran) Text(lang i18n.LanguageCodes, key string) string {
	return ""
}

func (m *MockTran) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return ""
}

func TestPodDistribution_ParsePodStatus(t *testing.T) {
	sdk := cptype.SDK{Tran: &MockTran{}}
	ctx := context.WithValue(context.Background(), cptype.GlobalInnerKeyCtxSDK, &sdk)
	pd := &PodDistribution{Data: Data{Total: 1}}
	status := pd.ParsePodStatus(ctx, "Running", 1)
	if status.Color != "green" {
		t.Errorf("test failed, value of status is unexpected")
	}
}

func TestPodDistribution_Transfer(t *testing.T) {
	component := &PodDistribution{
		Props: Props{
			RequestIgnore: []string{"test"},
		},
		Data: Data{
			Total: 1,
			Lists: []List{
				{
					Color: "green",
					Tip:   "1/1",
					Value: 1,
					Label: "1",
				},
			},
		},
	}
	c := &cptype.Component{}
	component.Transfer(c)
	ok, err := cputil.IsJsonEqual(c, component)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Errorf("test failed, json is not equal")
	}
}
