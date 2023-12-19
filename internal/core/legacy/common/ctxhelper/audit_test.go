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

package ctxhelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/i18n"
)

func TestAuditContext(t *testing.T) {

	type resp struct {
		language i18n.LanguageCodes
		flag     bool
	}
	testCase := []struct {
		name string
		lang any
		want resp
	}{
		{
			name: "invalid language",
			lang: 1,
			want: resp{
				language: nil,
				flag:     false,
			},
		},
		{
			name: "language is nil",
			lang: []*i18n.LanguageCode{},
			want: resp{
				language: nil,
				flag:     false,
			},
		},
		{
			name: "success get language",
			lang: i18n.LanguageCodes{
				&i18n.LanguageCode{
					Code:    en,
					Quality: 1,
				},
			},
			want: resp{
				language: i18n.LanguageCodes{
					&i18n.LanguageCode{
						Code:    en,
						Quality: 1,
					},
				},
				flag: true,
			},
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var lang i18n.LanguageCodes
			var flag bool
			if tt.lang == 1 {
				ctx := context.WithValue(context.Background(), languageKey, tt.lang)
				lang, flag = GetAuditLanguage(ctx)

			} else if tt.name == "language is nil" {
				lang, flag = GetAuditLanguage(context.Background())
			} else {
				ctx := PutAuditLanguage(context.Background(), tt.lang.(i18n.LanguageCodes))
				lang, flag = GetAuditLanguage(ctx)
			}
			assert.Equal(t, tt.want.language, lang)
			assert.Equal(t, tt.want.flag, flag)
		})
	}
}
