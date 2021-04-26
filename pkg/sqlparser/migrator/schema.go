// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package migrator

import (
	"fmt"

	"github.com/pingcap/parser/ast"
)

// Schema is the set of TableDefinitions.
// Is presents the status of the set of some tables.
type Schema struct {
	TableDefinitions map[string]*TableDefinition
}

func NewSchema() *Schema {
	return &Schema{TableDefinitions: make(map[string]*TableDefinition)}
}

func (s *Schema) Enter(in ast.Node) (ast.Node, bool) {
	switch in.(type) {
	case *ast.CreateTableStmt:
		tableStmt := in.(*ast.CreateTableStmt)
		s.TableDefinitions[tableStmt.Table.Name.String()] = NewTableDefinition(tableStmt)

	case *ast.DropTableStmt:
		for _, table := range in.(*ast.DropTableStmt).Tables {
			delete(s.TableDefinitions, table.Name.String())
		}

	case *ast.AlterTableStmt:
		alter := in.(*ast.AlterTableStmt)
		tableDefinition, ok := s.TableDefinitions[alter.Table.Name.String()]
		if ok {
			in.Accept(tableDefinition)
		}

	default:

	}

	return in, false
}

func (s *Schema) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (s *Schema) Equal(o *Schema) *Equal {
	if len(s.TableDefinitions) != len(o.TableDefinitions) {
		return &Equal{
			equal:  false,
			reason: fmt.Sprintf("left length: %v, right length: %v", len(s.TableDefinitions), len(o.TableDefinitions)),
		}
	}

	for tableName, sDef := range s.TableDefinitions {
		oDef, ok := o.TableDefinitions[tableName]
		if !ok {
			return &Equal{
				equal:  false,
				reason: fmt.Sprintf("table %s in left but not in right", tableName),
			}
		}
		if equal := sDef.Equal(oDef); !equal.Equal() {
			return equal
		}
	}

	return &Equal{equal: true}
}
