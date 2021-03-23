package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const varcharLengthLinterSQL = `
create table some_table (
	name varchar(5001)
);
`

func TestNewVarcharLengthLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewVarcharLengthLinter)
	if err := linter.Input([]byte(varcharLengthLinterSQL), "varcharLengthLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
