package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const charsetLinterSql = `
create table some_table (
 	public boolean
);

create table some_table (
	is_public int
) default charset = utf8mb4;
`

func TestNewCharsetLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewCharsetLinter)
	if err := linter.Input([]byte(charsetLinterSql), "charsetLinterSql"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) != 1 {
		t.Error("failed")
	}
}
