package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const tableCommentLinterTest = `
CREATE TABLE IF NOT EXISTS dice_api_slas(
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT 'primary key'
);

CREATE TABLE IF NOT EXISTS dice_api_slas(
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT 'primary key'
) comment='this is comment';
`

func TestNewTableCommentLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewTableCommentLinter)
	if err := linter.Input([]byte(tableCommentLinterTest), "tableCommentLinterTest"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
