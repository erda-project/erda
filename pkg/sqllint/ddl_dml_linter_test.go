package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const DDLDMLLinterSQL = `
begin;

update some_table set name='chen';

commit;

grant select on student to some_user with grant option;
`

func TestNewDDLDMLLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewDDLDMLLinter)
	if err := linter.Input([]byte(DDLDMLLinterSQL), "DDLDMLLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors["DDLDMLLinterSQL [lints]"]) != 3 {
		t.Fatal("failed")
	}
}
