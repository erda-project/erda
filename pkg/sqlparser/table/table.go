package table

import (
	"github.com/pingcap/parser/ast"
)

type Table struct {
	nodes []ast.DDLNode
}

func (t *Table) Append(node ast.DDLNode) {
	t.nodes = append(t.nodes, node)
}

func (t *Table) Nodes() []ast.DDLNode {
	return t.nodes
}
