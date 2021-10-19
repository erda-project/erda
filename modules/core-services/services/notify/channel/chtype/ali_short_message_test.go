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

package chtype

import "testing"

func TestAliShortMessage_Validate(t *testing.T) {
	type fields struct {
		AccessKeyId     string
		AccessKeySecret string
		SignName        string
		TemplateCode    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"case1", fields{AccessKeyId: "x", AccessKeySecret: "", SignName: "", TemplateCode: ""}, true},
		{"case2", fields{AccessKeyId: "", AccessKeySecret: "x", SignName: "", TemplateCode: ""}, true},
		{"case3", fields{AccessKeyId: "", AccessKeySecret: "", SignName: "x", TemplateCode: ""}, true},
		{"case4", fields{AccessKeyId: "", AccessKeySecret: "", SignName: "", TemplateCode: "x"}, true},
		{"case5", fields{AccessKeyId: "", AccessKeySecret: "", SignName: "", TemplateCode: ""}, true},
		{"case6", fields{AccessKeyId: "x", AccessKeySecret: "x", SignName: "x", TemplateCode: "x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asm := &AliShortMessage{
				AccessKeyId:     tt.fields.AccessKeyId,
				AccessKeySecret: tt.fields.AccessKeySecret,
				SignName:        tt.fields.SignName,
				TemplateCode:    tt.fields.TemplateCode,
			}
			if err := asm.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
