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

package i18nutil

import (
	"testing"

	"github.com/erda-project/erda-infra/providers/i18n"
)

func TestGetUserLang(t *testing.T) {
	type args struct {
		langs i18n.LanguageCodes
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "zh-CN",
			args: args{
				langs: i18n.LanguageCodes{
					{
						Code: "zh-CN",
					},
				},
			},
			want: Chinese,
		},
		{
			name: "zh",
			args: args{
				langs: i18n.LanguageCodes{
					{
						Code: "zh",
					},
				},
			},
			want: Chinese,
		},
		{
			name: "multi langs, use: zh",
			args: args{
				langs: i18n.LanguageCodes{
					{
						Code: "zh;q=0.9,en;q=0.8",
					},
				},
			},
			want: Chinese,
		},
		{
			name: "multi langs, use: en",
			args: args{
				langs: i18n.LanguageCodes{
					{
						Code: "zh;q=0.7,en;q=0.8",
					},
				},
			},
			want: Chinese,
		},
		{
			name: "unknown lang",
			args: args{
				langs: i18n.LanguageCodes{
					{
						Code: "jp;q=0.9,en;q=0.8",
					},
				},
			},
			want: Chinese,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserLang(tt.args.langs); got != tt.want {
				t.Errorf("GetUserLang() = %v, want %v", got, tt.want)
			}
		})
	}
}
