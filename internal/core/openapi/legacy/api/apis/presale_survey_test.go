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

package apis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validatePurveyDetail(t *testing.T) {
	type args struct {
		d *detail
		m map[string][]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nothing filled",
			args: args{
				d: &detail{},
				m: nil,
			},
			wantErr: true,
		},
		{
			name: "missing some fields",
			args: args{
				d: &detail{},
				m: map[string][]string{
					"realname": {"hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "all filled",
			args: args{
				d: &detail{
					RealName:    "",
					Mobile:      "",
					Email:       "",
					Position:    "",
					Company:     "",
					CompanySize: "",
					ITSize:      "",
					Purpose:     "",
				},
				m: map[string][]string{
					"realname":     {"hello"},
					"mobile":       {"mobile"},
					"email":        {"email"},
					"position":     {"position"},
					"company":      {"company"},
					"company_size": {"company_size"},
					"it_size":      {"it_size"},
					"purpose":      {"purpose"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePurveyDetail(tt.args.d, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("validatePurveyDetail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_makeDingdingMessage(t *testing.T) {
	d := detail{RealName: "hello"}
	content := makeDingdingMessage(d)
	assert.Contains(t, content, "真实姓名　　: hello")
}
