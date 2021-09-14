//  Copyright (c) 2021 Terminus, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package linters_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linters"
)

const nullToNotNullSQL1 = `
create table t1 (
	id bigint null
);

alter table t1 change id uuid bigint not null;
`

const nullToNotNullSQL2 = `
create table t1 (
	id bigint null
);

alter table t1 modify id bigint not null;
`

const nullToNotNullSQL3 = `
create table t1 (
	id bigint not null
);

alter table t1 change id uuid bigint not null;
`

const nullToNotNullSQL4 = `
create table t1 (
	id bigint not null
);

alter table t1 modify id bigint not null;
`

const nullToNotNullSQL5 = `
create table t1 (
	id bigint not null
);

alter table t1 modify id bigint null  ;
`

const nullToNotNullSQL6 = `
create table t1 (
	id bigint
);

alter table t1 modify id bigint not null;
`

type nullToNotNullLinterCase struct {
	sql      string
	filename string
	hasError bool
}

func TestNewNullToNotNullLinter(t *testing.T) {
	for _, case_ := range []nullToNotNullLinterCase{
		{
			sql:      nullToNotNullSQL1,
			filename: "nullToNotNullSQL1",
			hasError: true,
		}, {
			sql:      nullToNotNullSQL2,
			filename: "nullToNotNullSQL2",
			hasError: true,
		}, {
			sql:      nullToNotNullSQL3,
			filename: "nullToNotNullSQL3",
			hasError: false,
		}, {
			sql:      nullToNotNullSQL4,
			filename: "nullToNotNullSQL4",
			hasError: false,
		}, {
			sql:      nullToNotNullSQL5,
			filename: "nullToNotNullSQL5",
			hasError: false,
		}, {
			sql:      nullToNotNullSQL6,
			filename: "nullToNotNullSQL6",
			hasError: false,
		},
	} {
		linter := sqllint.New(linters.NewNullToNotNullLinter)
		if err := linter.Input([]byte(case_.sql), case_.filename); err != nil {
			t.Fatalf("filename: %s: %v", case_.filename, err)
		}
		linter.FprintErrors(nil)
		if linter.HasError() != case_.hasError {
			t.Fatalf("assert error, filename: %s", case_.filename)
		}
	}
}
