package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const tablenameLinterSQL = `
create table some_0_table (
	id int
);

create table Some_table (
	id int
);

create table 某表 (
	id int
);
`

func TestNewTableNameLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewTableNameLinter)
	if err := linter.Input([]byte(tablenameLinterSQL), "tablenameLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
