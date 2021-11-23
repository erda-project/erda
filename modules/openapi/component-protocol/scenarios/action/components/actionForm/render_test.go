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

package action

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/i18n"
)

func TestGenTimeoutProps1(t *testing.T) {
	type args struct {
		local *i18n.LocaleResource
	}
	tests := []struct {
		name      string
		args      args
		wantProps []apistructs.FormPropItem
		wantErr   bool
	}{
		{
			name: "test",
			args: args{
				local: &i18n.LocaleResource{},
			},
			wantProps: []apistructs.FormPropItem{
				{
					Label:     "wb.content.action.input.label.timeout",
					Component: "inputNumber",
					Key:       "timeout",
					ComponentProps: map[string]interface{}{
						"placeholder": "wb.content.action.input.label.timeoutPlaceholder",
					},
					DefaultValue: "3600",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProps, err := GenTimeoutProps(tt.args.local)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenTimeoutProps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotProps, tt.wantProps) {
				t.Errorf("GenTimeoutProps() gotProps = %v, want %v", gotProps, tt.wantProps)
			}
		})
	}
}
