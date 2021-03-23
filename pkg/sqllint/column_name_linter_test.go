package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const columnNameLinterSQL = `
create table some_table (
	姓名 varchar(100)
);

create table some_table (
	BigBang varchar(100)
);

create table some_table (
	3d_school varchar(100)
);

create table some_table (
	level_3_vip varchar(100)
);
`

func TestNewColumnNameLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewColumnNameLinter)
	if err := linter.Input([]byte(columnNameLinterSQL), "columnNameLinterSQL"); err != nil {
		t.Error(err)
	}

	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors["columnNameLinterSQL [lints]"]) != 4 {
		t.Fatal("failed", len(errors["columnNameLinterSQL"]))
	}
}
