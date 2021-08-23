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

package linters_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linters"
)

const destructLinterSLQ = `
drop table some_table;

drop database some_base;

drop user some_user;

truncate table some_table;

rename table dice_api_access to dice_api_access2;

alter table dice_api_access
    drop column org_id;

alter table dice_api_access 
	change asset_id asset_id2 varchar(191) null comment 'asset id';
`

func TestNewDestructLinter(t *testing.T) {
	linter := sqllint.New(linters.NewDestructLinter)
	if err := linter.Input([]byte(destructLinterSLQ), "destructLinterSLQ"); err != nil {
		t.Error(err)
	}

	errors := linter.Errors()
	t.Logf("errors: %v", errors)

}
