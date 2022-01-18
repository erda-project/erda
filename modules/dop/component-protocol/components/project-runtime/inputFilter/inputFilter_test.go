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

package page

import (
	"fmt"
	"testing"

	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/i18n"
)

type NopTranslator struct{}

func (t NopTranslator) Get(lang i18n.LanguageCodes, key, def string) string { return key }

func (t NopTranslator) Text(lang i18n.LanguageCodes, key string) string { return key }

func (t NopTranslator) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return fmt.Sprintf(key, args...)
}

var defaultSDK = &cptype.SDK{

	GlobalState: &cptype.GlobalStateData{},
	Tran:        &NopTranslator{},
}

func TestInputFilter_getOperation(t *testing.T) {
	type fields struct {
		DefaultProvider base.DefaultProvider
		Type            string
		sdk             *cptype.SDK
		State           State
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			name: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &InputFilter{}
			p.getOperation()
		})
	}
}

func TestInputFilter_getState(t *testing.T) {

	type args struct {
		sdk *cptype.SDK
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{sdk: defaultSDK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &InputFilter{}
			p.getState(tt.args.sdk, &cptype.Component{})
		})
	}
}

func TestInputFilter_flushOptsByFilter(t *testing.T) {
	type args struct {
		filterEntity string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{filterEntity: "×"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &InputFilter{}
			af.flushOptsByFilter(tt.args.filterEntity)
		})
	}
}

func TestInputFilter_generateUrlQueryParams(t *testing.T) {
	type args struct {
		Values cptype.ExtraMap
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1",
			args: args{Values: map[string]interface{}{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			af := &InputFilter{}
			af.generateUrlQueryParams()
		})
	}
}
