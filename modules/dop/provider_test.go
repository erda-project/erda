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

package dop

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/protobuf/proto-go/cp/pb"
	"github.com/erda-project/erda-infra/providers/i18n"
)

type MockCP struct {
	Tran i18n.I18n
}

func (m *MockCP) Render(context.Context, *pb.RenderRequest) (*pb.RenderResponse, error) {
	return nil, nil
}
func (m *MockCP) SetI18nTran(tran i18n.I18n)              { m.Tran = tran }
func (m *MockCP) WithContextValue(key, value interface{}) {}

type MockI18n struct{}

func (m *MockI18n) Get(namespace string, lang i18n.LanguageCodes, key, def string) string { return "" }
func (m *MockI18n) Text(namespace string, lang i18n.LanguageCodes, key string) string     { return "" }
func (m *MockI18n) Sprintf(namespace string, lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return ""
}
func (m MockI18n) Translator(namespace string) i18n.Translator { return &i18n.NopTranslator{} }

func Test_provider_Init(t *testing.T) {

	mockCP := &MockCP{}
	nopTran := &MockI18n{}
	p := &provider{Log: logrusx.New(), Protocol: mockCP, CPTran: nopTran}
	monkey.PatchInstanceMethod(reflect.TypeOf(p), "Initialize",
		func(p *provider, ctx servicehub.Context) error {
			return nil
		})
	defer monkey.UnpatchAll()

	err := p.Init(nil)
	assert.NoError(t, err)
	assert.Equal(t, mockCP.Tran, nopTran)
}
