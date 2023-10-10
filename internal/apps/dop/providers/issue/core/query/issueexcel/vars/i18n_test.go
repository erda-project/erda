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

package vars

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/i18n"
)

func TestParseLangHeader(t *testing.T) {
	lang, err := i18n.ParseLanguageCode(I18nLang_zh_CN)
	assert.NoError(t, err)
	for _, l := range lang {
		assert.Equal(t, I18nLang_zh_CN, l.Code)
		assert.Equal(t, float32(1), l.Quality)
	}
}
