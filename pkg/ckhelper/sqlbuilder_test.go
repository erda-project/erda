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

package ckhelper

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	testCases := []struct {
		name    string
		sel     interface{}
		wantsql string
	}{
		{
			name:    "FromTimestampMilli",
			sel:     FromTimestampMilli(1660185863000),
			wantsql: "SELECT fromUnixTimestamp64Milli(toInt64(1660185863000))",
		},
		{
			name:    "FromTagsKey",
			sel:     FromTagsKey("tags.org_name"),
			wantsql: "SELECT tag_values[indexOf(tag_keys, 'org_name')]",
		},
		{
			name:    "FromFieldNumberKey",
			sel:     FromFieldNumberKey("fields.read_bytes"),
			wantsql: "SELECT number_field_values[indexOf(number_field_keys, 'read_bytes')]",
		},
		{
			name:    "FromFieldStringKey",
			sel:     FromFieldStringKey("fields.level"),
			wantsql: "SELECT string_field_values[indexOf(field_string_keys, 'level')]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := goqu.From()
			sql, _, err := ds.Select(tc.sel).ToSQL()
			if err != nil {
				t.Errorf("tosql must successful: %s", err)
				return
			}
			assert.Equal(t, tc.wantsql, sql)
		})
	}
}
