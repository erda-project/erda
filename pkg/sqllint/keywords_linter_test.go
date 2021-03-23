package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const keywordsLinterSQL = `
ALTER TABLE pipeline_labels
    CHANGE COLUMN value value varchar(191) NOT NULL DEFAULT '' COMMENT '标签值';
`

func TestNewKeywordsLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewKeywordsLinter)
	if err := linter.Input([]byte(keywordsLinterSQL), "keywordsLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
