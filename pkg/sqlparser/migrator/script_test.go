package migrator_test

import (
	"testing"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/sqlparser/migrator"
)

func TestNewScript(t *testing.T) {
	script, err := migrator.NewScript("..", "testdata/dice_base.sql")
	if err != nil {
		t.Fatal(err)
	}

	for _, dml := range script.DMLNodes() {
		t.Log(dml.Text())
		stmt, ok := dml.(*ast.SetStmt)
		if ok {
			t.Log(stmt.Text())
		}
	}
}
