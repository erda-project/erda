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

package snapshot

import (
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/sqlparser/table"
)

type Snapshot struct {
	tables map[string]*table.Table
}

func From(tx *gorm.DB, ignore ...string) (s *Snapshot, err error) {
	s = &Snapshot{tables: make(map[string]*table.Table, 0)}
	ignores := make(map[string]bool, len(ignore))
	for _, ig := range ignore {
		ignores[ig] = true
	}

	defer func() {
		if err != nil {
			err = errors.Wrap(err, "failed to snapshot database schema structure")
		}
	}()

	tx = tx.Raw("SHOW TABLES")
	if err = tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s, nil
		}
		return nil, err
	}

	rows, err := tx.Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			tableName string
			stmt      string
		)
		if err = rows.Scan(&tableName); err != nil {
			return nil, err
		}
		if _, ok := ignores[tableName]; ok {
			continue
		}

		tx := tx.Raw("SHOW CREATE TABLE " + tableName)
		if err = tx.Error; err != nil {
			return nil, err
		}
		if err = tx.Row().Scan(&tableName, &stmt); err != nil {
			return nil, err
		}
		node, err := parser.New().ParseOneStmt(stmt, "", "")
		if err != nil {
			return nil, err
		}
		t := new(table.Table)
		t.Append(node.(ast.DDLNode))
		s.tables[tableName] = t
	}

	return s, err
}

func (s *Snapshot) DDLNodes() []ast.DDLNode {
	var nodes []ast.DDLNode
	for _, t := range s.tables {
		nodes = append(nodes, t.Nodes()...)
	}
	return nodes
}

func (s *Snapshot) HasAnyTable() bool {
	return len(s.tables) > 0
}

func (s *Snapshot) TableNames() []string {
	var names []string
	for k := range s.tables {
		names = append(names, k)
	}
	return names
}

func (s *Snapshot) RecoverTo(tx *gorm.DB) error {
	nodes := s.DDLNodes()

	var (
		installing = make(map[string]*ast.CreateTableStmt, 0)
		installed  = make(map[string]*ast.CreateTableStmt, 0)
	)

	for _, ddl := range nodes {
		create := ddl.(*ast.CreateTableStmt)
		installing[create.Table.Name.String()] = create
	}

	// f is used to execute every CreateTableStmt .
	// f can resolve foreign key dependencies recursively.
	var f func(stmt *ast.CreateTableStmt) error
	f = func(create *ast.CreateTableStmt) error {
		// skip installed table
		if _, ok := installed[create.Table.Name.String()]; ok {
			return nil
		}

		// resolve reference dependencies
		for _, constraint := range create.Constraints {
			if constraint.Tp == ast.ConstraintForeignKey {
				if alia, ok := installing[constraint.Refer.Table.Name.String()]; ok {
					if err := f(alia); err != nil {
						return err
					}
				}
			}
		}

		// install
		if err := tx.Exec(create.Text()).Error; err != nil {
			return err
		}
		installed[create.Table.Name.String()] = create
		return nil
	}

	for _, create := range installing {
		if err := f(create); err != nil {
			return err
		}
	}

	return nil
}
