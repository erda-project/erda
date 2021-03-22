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
