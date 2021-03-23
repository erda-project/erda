package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const booleanFieldLinterSQL = `
create table some_table (
 	public boolean
);

create table some_table (
	public tinyint(1)
);

create table some_table (
	is_public int
);
`

func TestNewBooleanFieldLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewBooleanFieldLinter)
	if err := linter.Input([]byte(booleanFieldLinterSQL), "booleanFieldLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) != 1 {
		t.Error("failed")
	}
}
