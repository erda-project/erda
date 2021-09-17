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

package issuestream

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func Test_getIssueStreamTemplate(t *testing.T) {
	type args struct {
		locale string
		ist    apistructs.IssueStreamType
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "zh exist",
			args: args{
				locale: "zh",
				ist:    apistructs.ISTChangeIteration,
			},
			want:    apistructs.IssueTemplate["zh"][apistructs.ISTChangeIteration],
			wantErr: false,
		},
		{
			name: "en exist",
			args: args{
				locale: "en",
				ist:    apistructs.ISTComment,
			},
			want:    apistructs.IssueTemplate["zh"][apistructs.ISTComment],
			wantErr: false,
		},
		{
			name: "zh not exist",
			args: args{
				locale: "zh",
				ist:    "not exist",
			},
			wantErr: true,
		},
		{
			name: "en not exist",
			args: args{
				locale: "en",
				ist:    "not exist",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIssueStreamTemplate(tt.args.locale, tt.args.ist)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIssueStreamTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIssueStreamTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIssueStreamTemplateForMsgSending(t *testing.T) {
	type args struct {
		locale string
		ist    apistructs.IssueStreamType
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "zh override exist",
			args: args{
				locale: "zh",
				ist:    apistructs.ISTComment,
			},
			want:    apistructs.IssueTemplateOverrideForMsgSending["zh"][apistructs.ISTComment],
			wantErr: false,
		},
		{
			name: "zh exist, but override not exist",
			args: args{
				locale: "zh",
				ist:    apistructs.ISTChangeIteration,
			},
			want:    apistructs.IssueTemplate["zh"][apistructs.ISTChangeIteration],
			wantErr: false,
		},
		{
			name: "zh not exist",
			args: args{
				locale: "zh",
				ist:    "not exist",
			},
			wantErr: true,
		},
		{
			name: "en override exist",
			args: args{
				locale: "en",
				ist:    apistructs.ISTComment,
			},
			want:    apistructs.IssueTemplateOverrideForMsgSending["en"][apistructs.ISTComment],
			wantErr: false,
		},
		{
			name: "en exist, but override not exist",
			args: args{
				locale: "en",
				ist:    apistructs.ISTChangeIteration,
			},
			want:    apistructs.IssueTemplate["en"][apistructs.ISTChangeIteration],
			wantErr: false,
		},
		{
			name: "en not exist",
			args: args{
				locale: "en",
				ist:    "not exist",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIssueStreamTemplateForMsgSending(tt.args.locale, tt.args.ist)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIssueStreamTemplateForMsgSending() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIssueStreamTemplateForMsgSending() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDefaultContent(t *testing.T) {
	type args struct {
		ist   apistructs.IssueStreamType
		param apistructs.ISTParam
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "comment",
			args: args{
				ist:   apistructs.ISTComment,
				param: apistructs.ISTParam{Comment: "hello world"},
			},
			want:    `hello world`,
			wantErr: false,
		},
		{
			name: "change content",
			args: args{
				ist: apistructs.ISTChangeContent,
				param: apistructs.ISTParam{
					CurrentContent: "old",
					NewContent:     "new",
				},
			},
			want:    `该事件内容发生变更`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDefaultContent(tt.args.ist, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDefaultContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDefaultContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDefaultContentForMsgSending(t *testing.T) {
	type args struct {
		ist   apistructs.IssueStreamType
		param apistructs.ISTParam
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "comment",
			args: args{
				ist: apistructs.ISTComment,
				param: apistructs.ISTParam{
					Comment: "hello world",
				},
			},
			want:    `添加了备注: hello world`,
			wantErr: false,
		},
		{
			name: "change content",
			args: args{
				ist: apistructs.ISTChangeContent,
			},
			want:    apistructs.IssueTemplate["zh"][apistructs.ISTChangeContent],
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDefaultContentForMsgSending(tt.args.ist, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDefaultContentForMsgSending() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDefaultContentForMsgSending() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renderTemplate(t *testing.T) {
	type args struct {
		locale          string
		templateContent string
		param           apistructs.ISTParam
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "change iteration",
			args: args{
				locale:          "zh",
				templateContent: apistructs.IssueTemplate["zh"][apistructs.ISTChangeIteration],
				param:           apistructs.ISTParam{CurrentIteration: "1.2", NewIteration: "1.3"},
			},
			want:    `该事件迭代由 "1.2" 变更为 "1.3"`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate(tt.args.locale, tt.args.templateContent, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("renderTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
