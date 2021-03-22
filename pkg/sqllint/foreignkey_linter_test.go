package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const foreignkeyLinterSQL = `
ALTER TABLE students
ADD CONSTRAINT fk_class_id
FOREIGN KEY (class_id)
REFERENCES classes (id);
`

func TestNewForeignKeyLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewForeignKeyLinter)
	if err := linter.Input([]byte(foreignkeyLinterSQL), "foreignkeyLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
