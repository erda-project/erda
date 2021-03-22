package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const floatLinterSQL = `
create table some_table (
	-- score float,
	score_rate double
);
`

func TestNewFloatDoubleLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewFloatDoubleLinter)
	if err := linter.Input([]byte(floatLinterSQL), "floatLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors["floatLinterSQL [lints]"]) == 0 {
		t.Fatal("failed")
	}
}
