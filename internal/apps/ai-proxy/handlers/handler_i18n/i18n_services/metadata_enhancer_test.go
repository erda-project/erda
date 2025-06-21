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

package i18n_services

import (
	"testing"
)

func TestGetLocaleFromContext(t *testing.T) {
	type args struct {
		inputLang string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// Test case 1: Multi-language with quality values - English priority
		{
			name: "multi-language-with-quality-english-priority",
			args: args{
				inputLang: "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
			},
			want: "en",
		},
		// Test case 2: Multi-language with quality values - Chinese priority
		{
			name: "multi-language-with-quality-chinese-priority",
			args: args{
				inputLang: "zh-CN,zh;q=0.9,en;q=0.8,ja;q=0.7",
			},
			want: "zh",
		},
		// Test case 3: Multi-language with quality values - Japanese priority
		{
			name: "multi-language-with-quality-japanese-priority",
			args: args{
				inputLang: "ja,en-US;q=0.9,zh-CN;q=0.8",
			},
			want: "ja",
		},
		// Test case 4: Multi-language with quality values - French priority with region code
		{
			name: "multi-language-with-quality-french-priority-with-region",
			args: args{
				inputLang: "fr-FR;q=1.0,en-US;q=0.9,zh-CN;q=0.8",
			},
			want: "fr",
		},
		// Test case 5: Single language - English
		{
			name: "single-language-english",
			args: args{
				inputLang: "en",
			},
			want: "en",
		},
		// Test case 6: Single language - Chinese with region code
		{
			name: "single-language-chinese-with-region",
			args: args{
				inputLang: "zh-CN",
			},
			want: "zh",
		},
		// Test case 7: Single language - Japanese
		{
			name: "single-language-japanese",
			args: args{
				inputLang: "ja",
			},
			want: "ja",
		},
		// Test case 8: No Accept-Language header
		{
			name: "no-accept-language-header",
			args: args{
				inputLang: "",
			},
			want: "zh",
		},
		// Test case 9: Empty Accept-Language header
		{
			name: "empty-accept-language-header",
			args: args{
				inputLang: "",
			},
			want: "zh",
		},
		// Test case 10: Invalid Accept-Language format
		{
			name: "invalid-accept-language-format",
			args: args{
				inputLang: "invalid-format",
			},
			want: "zh",
		},
		// Test case 11: Complex quality values - German highest
		{
			name: "complex-quality-values-german-highest",
			args: args{
				inputLang: "de;q=1.0,en-US;q=0.9,fr;q=0.8,zh-CN;q=0.7,ja;q=0.6",
			},
			want: "de",
		},
		// Test case 12: Same quality values - First one priority
		{
			name: "same-quality-values-first-priority",
			args: args{
				inputLang: "es;q=0.9,it;q=0.9,pt;q=0.9",
			},
			want: "es",
		},
		// Test case 13: Complex region codes
		{
			name: "complex-region-codes",
			args: args{
				inputLang: "en-GB-oed,en-US;q=0.9,zh-Hans-CN;q=0.8",
			},
			want: "en",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLocaleFromContext(tt.args.inputLang); got != tt.want {
				t.Errorf("GetLocaleFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
