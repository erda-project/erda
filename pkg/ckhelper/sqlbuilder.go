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
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"strings"
)

const (
	tagsPrefix   = "tags."
	fieldsPrefix = "fields."
)

// FromTimestampMilli convert timestamp(ms) to Datetime type in clickhouse
func FromTimestampMilli(ts int64) exp.SQLFunctionExpression {
	return goqu.Func("fromUnixTimestamp64Milli", goqu.Func("toInt64", ts))
}

// FromTagsKey convert tags.key_1 to tag_values[indexOf(tag_keys, 'key')]
func FromTagsKey(key string) exp.LiteralExpression {
	return goqu.L("tag_values[indexOf(tag_keys, ?)]", TrimTags(key))
}

func FromFieldNumberKey(key string) exp.LiteralExpression {
	return goqu.L("number_field_values[indexOf(number_field_keys, ?)]", TrimFields(key))
}

func FromFieldStringKey(key string) exp.LiteralExpression {
	return goqu.L("string_field_values[indexOf(field_string_keys, ?)]", TrimFields(key))
}

func TrimTags(key string) string {
	return strings.TrimPrefix(key, tagsPrefix)
}

func TrimFields(field string) string {
	return strings.TrimPrefix(field, fieldsPrefix)
}
