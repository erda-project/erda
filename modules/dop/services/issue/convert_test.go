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

package issue

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/magiconair/properties/assert"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/i18n"
)

func TestIssue_getIssueExportDataI18n(t *testing.T) {
	bdl := bundle.New(bundle.WithI18nLoader(&i18n.LocaleResourceLoader{}))
	m := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetLocale",
		func(bdl *bundle.Bundle, local ...string) *i18n.LocaleResource {
			return &i18n.LocaleResource{}
		})
	defer m.Unpatch()

	svc := New(WithBundle(bdl))
	strs := svc.getIssueExportDataI18n("testKey", "test")
	assert.Equal(t, strs, []string{"test"})
}
