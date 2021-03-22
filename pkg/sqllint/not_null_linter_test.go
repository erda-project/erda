package sqllint_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const notNullLinterTest = `
create table some_table (
	name varchar(101)
);

ALTER TABLE dice_api_access
    ADD default_sla_id BIGINT COMMENT 'default SLA id';
`

func TestNewNotNullLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewNotNullLinter)
	if err := linter.Input([]byte(notNullLinterTest), "notNullLinterTest"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}

	scanner := bufio.NewScanner(bytes.NewBufferString(notNullLinterTest))
	var firstLine string
	if scanner.Scan() {
		firstLine = scanner.Text()
	}
	t.Logf("firstLine: %s", firstLine)
	for scanner.Scan() {
		t.Log(scanner.Text())
	}
}
