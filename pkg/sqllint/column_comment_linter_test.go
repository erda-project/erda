package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const columnCommentLinterSQL = `
create table some_table (
	name varchar(101)
);

alter table some_table
	add name varchar(101)
;
`

func TestNewCommentLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewColumnCommentLinter)
	if err := linter.Input([]byte(columnCommentLinterSQL), "columnCommentLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors["columnCommentLinterSQL [lints]"]) != 2 {
		t.Fatal("failed")
	}
}
