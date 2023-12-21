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

	"github.com/erda-project/erda-infra/providers/i18n"
)

const languageKey = "language"
const zh = "zh-CN"
const en = "en-US"

func GetAuditLanguage(ctx context.Context) (i18n.LanguageCodes, bool) {
	langCodes, ok := ctx.Value(languageKey).(i18n.LanguageCodes)
	if !ok {
		return nil, false
	}
	if langCodes == nil {
		return i18n.LanguageCodes{
			&i18n.LanguageCode{
				Code:    zh,
				Quality: 1,
			},
		}, true
	}
	return langCodes, true
}

func PutAuditLanguage(ctx context.Context, language i18n.LanguageCodes) context.Context {
	return context.WithValue(ctx, languageKey, language)
}
