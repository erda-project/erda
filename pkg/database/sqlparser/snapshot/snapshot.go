// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package snapshot

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/database/sqlparser/table"
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
		stmt = TrimCharacterSetFromRawCreateTableSQL(stmt)
		stmt = TrimBlockFormat(stmt)
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
		var (
			buf = bytes.NewBuffer(nil)
		)
		TrimCollateOptionFromCreateTable(create)
		TrimCollateOptionFromCols(create)
		TrimConstraintCheckFromCreateTable(create)

		if err := create.Restore(&format.RestoreCtx{
			Flags:     format.DefaultRestoreFlags,
			In:        buf,
			JoinLevel: 0,
		}); err != nil {
			return errors.Wrap(err, "failed to Restore table definition")
		}

		if err := tx.Exec(buf.String()).Error; err != nil {
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

func TrimCollateOptionFromCols(createStmt *ast.CreateTableStmt) {
	if createStmt == nil {
		return
	}
	for i := range createStmt.Cols {
		for j := len(createStmt.Cols[i].Options) - 1; j >= 0; j-- {
			if createStmt.Cols[i].Options[j].Tp == ast.ColumnOptionCollate {
				createStmt.Cols[i].Options = append(createStmt.Cols[i].Options[:j], createStmt.Cols[i].Options[j+1:]...)
			}
		}
	}
}

func TrimCollateOptionFromCreateTable(createStmt *ast.CreateTableStmt) {
	if createStmt == nil {
		return
	}
	for i := len(createStmt.Options) - 1; i >= 0; i-- {
		if createStmt.Options[i].Tp == ast.TableOptionCollate {
			createStmt.Options = append(createStmt.Options[:i], createStmt.Options[i+1:]...)
		}
	}
}

func TrimConstraintCheckFromCreateTable(createStmt *ast.CreateTableStmt) {
	if createStmt == nil {
		return
	}
	for i := len(createStmt.Constraints) - 1; i >= 0; i-- {
		if createStmt.Constraints[i].Tp == ast.ConstraintCheck {
			createStmt.Constraints = append(createStmt.Constraints[:i], createStmt.Constraints[i+1:]...)
		}
	}
}

func TrimCharacterSetFromRawCreateTableSQL(createStmt string) string {
	return regexp.MustCompile(`(?i)(?:DEFAULT)* (?:CHARACTER SET|CHARSET)\s*=\s*\w+`).ReplaceAllString(createStmt, "")
}

func TrimBlockFormat(createStmt string) string {
	return strings.ReplaceAll(createStmt, "BLOCK_FORMAT=ENCRYPTED", "")
}

// ParseCreateTableStmt parses CreateTableStmt as *ast.CreateTableStmt node
func ParseCreateTableStmt(createStmt string) (*ast.CreateTableStmt, error) {
	createStmt = TrimCharacterSetFromRawCreateTableSQL(createStmt)
	createStmt = TrimBlockFormat(createStmt)
	node, err := parser.New().ParseOneStmt(createStmt, "", "")
	if err != nil {
		return nil, err
	}
	createTableStmt, ok := node.(*ast.CreateTableStmt)
	if !ok {
		return nil, errors.Errorf("the text is not CreateTableStmt, text: %s", createStmt)
	}
	return createTableStmt, nil
}
