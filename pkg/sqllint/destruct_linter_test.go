// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
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
	linter := sqllint.New(sqllint.NewDestructLinter)
	if err := linter.Input([]byte(destructLinterSLQ), "destructLinterSLQ"); err != nil {
		t.Error(err)
	}

	errors := linter.Errors()
	t.Logf("errors: %v", errors)

}
