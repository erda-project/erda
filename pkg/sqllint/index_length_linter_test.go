package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const indexLengthLinterSQL = `
create table some_table (
	name varchar(101),
	index idx_name (name(200))
);

create table some_table (
	some_text varchar(200),
	index idx_some_text (some_text(100))
);

create table some_table (
	some_text varchar(200),
	index idx_some_text (some_text)
);

create table some_table (
	name varchar(300),
	some_text varchar(500),
	index idx_name (name, some_text)
);
`

func TestNewIndexLengthLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewIndexLengthLinter)
	if err := linter.Input([]byte(indexLengthLinterSQL), "indexLengthLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
