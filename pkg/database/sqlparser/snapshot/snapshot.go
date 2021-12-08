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
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/database/sqlparser/table"
)

var (
	charsetWhiteEnv     = "MIGRATION_CHARSET_WHITE"
	defaultCharsetWhite = []string{"utf8", "utf8mb4"}
)

// Snapshot maintains the structure of the database tables
type Snapshot struct {
	tables map[string]*table.Table
	from   *gorm.DB
}

// From snapshots the structure of the database tables, and returns the Snapshot.
// tx is the connection handler of the goal DB.
// ignore is the tables you do not want to snapshot.
func From(tx *gorm.DB, ignore ...string) (s *Snapshot, err error) {
	s = &Snapshot{tables: make(map[string]*table.Table), from: tx}
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
		// some character set can not be parsed, so trim them from the stmt
		stmt = TrimCharacterSetFromRawCreateTableSQL(stmt, CharsetWhite()...)
		// block format syntax can not be parsed, so trim it
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

// DDLNodes returns DDLs
func (s *Snapshot) DDLNodes() []ast.DDLNode {
	var nodes []ast.DDLNode
	for _, t := range s.tables {
		nodes = append(nodes, t.Nodes()...)
	}
	return nodes
}

// HasAnyTable returns true if there is any table in the Snapshot
func (s *Snapshot) HasAnyTable() bool {
	return len(s.tables) > 0
}

// TableNames returns all tables names
func (s *Snapshot) TableNames() []string {
	var names []string
	for k := range s.tables {
		names = append(names, k)
	}
	return names
}

// Dump dumps data from DB.
// tableName is the table you want to dump.
// lines is the numbers of lines you want to dump.
// Note: lines is an approximate number rather than an exact number.
// Note: the max data size is not more than 1<<16 (the MySQL default max placeholders size).
func (s *Snapshot) Dump(tableName string, lines uint64) ([]map[string]interface{}, uint64, error) {
	var (
		selectCount = fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
		selectAll   = fmt.Sprintf("SELECT * FROM `%s` ORDER BY RAND() LIMIT %d", tableName, lines)
		count       uint64
	)
	tx := s.from.Raw(selectCount)
	if err := tx.Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed to Raw(%s)", strconv.Quote(selectCount))
	}
	if err := tx.Row().Scan(&count); err != nil {
		return nil, 0, errors.Wrapf(err, "failed to Scan count from %s", strconv.Quote(selectCount))
	}
	if count == 0 {
		return nil, count, nil
	}
	if err := tx.Raw(selectAll).Error; err != nil {
		return nil, count, errors.Wrapf(err, "failed to Raw(%s)", strconv.Quote(selectAll))
	}
	rows, err := tx.Rows()
	if err != nil {
		return nil, count, errors.Wrapf(err, "failed to Rows from %s", strconv.Quote(selectAll))
	}
	defer rows.Close()

	var data []map[string]interface{}
	for rows.Next() {
		columns, _ := rows.Columns()
		values := make([]interface{}, len(columns))
		for i := 0; i < len(values); i++ {
			values[i] = &values[i]
		}
		if err := rows.Scan(values...); err != nil {
			return nil, count, errors.Wrapf(err, "failed to Scan %v, from %s", columns, strconv.Quote(selectAll))
		}
		var record = make(map[string]interface{})
		for i := 0; i < len(columns); i++ {
			record[columns[i]] = values[i]
		}
		data = append(data, record)
		// (1<<16)-1 is the max placeholders size to insert to for mysql default
		if (len(data)+1)*len(columns) >= 1<<16 {
			break
		}
	}
	return data, count, nil
}

// RecoverTo recover the data structure and data to goal db
func (s *Snapshot) RecoverTo(tx *gorm.DB) error {
	var (
		l          = logrus.WithField("func", "*Snapshot.RecoverTo")
		nodes      = s.DDLNodes()
		installing = make(map[string]*ast.CreateTableStmt)
		installed  = make(map[string]*ast.CreateTableStmt)
		inserted   = make(map[string]*ast.CreateTableStmt)
	)

	for _, ddl := range nodes {
		create := ddl.(*ast.CreateTableStmt)
		installing[create.Table.Name.String()] = create
	}

	var (
		// createF is used to execute every CreateTableStmt .
		// createF can resolve foreign key dependencies recursively.
		createF func(*ast.CreateTableStmt) error
		// insertF dumps data and inserts into new db.
		insertF func(*ast.CreateTableStmt, uint64) (int, uint64, error)
	)
	createF = func(create *ast.CreateTableStmt) error {
		// skip installed table
		if _, ok := installed[create.Table.Name.String()]; ok {
			return nil
		}

		// resolve reference dependencies
		for _, constraint := range create.Constraints {
			if constraint.Tp == ast.ConstraintForeignKey {
				if alia, ok := installing[constraint.Refer.Table.Name.String()]; ok {
					if err := createF(alia); err != nil {
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
	insertF = func(create *ast.CreateTableStmt, lines uint64) (int, uint64, error) {
		// skip inserted table
		if _, ok := inserted[create.Table.Name.String()]; ok {
			return 0, 0, nil
		}

		// resolve reference dependencies
		exLines := lines * 5
		for _, constraint := range create.Constraints {
			if constraint.Tp == ast.ConstraintForeignKey {
				if alia, ok := installing[constraint.Refer.Table.Name.String()]; ok {
					lines = exLines
					if n, count, err := insertF(alia, lines); err != nil {
						return n, count, err
					}
				}
			}
		}

		// insert
		inserted[create.Table.Name.String()] = create
		data, count, err := s.Dump(create.Table.Name.String(), lines)
		if err != nil {
			return 0, count, err
		}
		n := len(data)
		if n == 0 {
			return n, count, nil
		}
		if err = tx.Table(create.Table.Name.String()).Create(data).Error; err == nil {
			return n, count, nil
		}
		l.WithError(err).
			WithField("func", "*Snapshot.RecoverTo.insertF").
			WithField("tableName", create.Table.Name.String()).
			Warnln("failed to bulk insert, try to insert one by one item")
		n = 0
		for _, item := range data {
			if err := tx.Table(create.Table.Name.String()).Create(item).Error; err != nil {
				if IsCannotAddOrUpdateAChildRowError(err) {
					continue
				}
				return n, count, err
			}
			n++
		}

		return n, count, nil
	}

	for _, create := range installing {
		if err := createF(create); err != nil {
			return err
		}
		if Sampling() {
			n, count, err := insertF(create, MaxSampling())
			l.WithField("tableName", create.Table.Name.String()).Infof("collect %d/%d (sampling/total) lines", n, count)
			if err != nil {
				return err
			}
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

func TrimCharacterSetFromRawCreateTableSQL(createStmt string, except ...string) string {
	pat := `(?i)(?:DEFAULT)* (?:CHARACTER SET|CHARSET)\s*=\s*(\w+)`
	re := regexp.MustCompile(pat)
	found := re.FindStringSubmatch(createStmt)
	if len(found) == 0 {
		return createStmt
	}
	for _, ex := range except {
		if strings.EqualFold(ex, found[len(found)-1]) {
			return createStmt
		}
	}
	return re.ReplaceAllString(createStmt, "")
}

func TrimBlockFormat(createStmt string) string {
	return strings.ReplaceAll(createStmt, "BLOCK_FORMAT=ENCRYPTED", "")
}

// ParseCreateTableStmt parses CreateTableStmt as *ast.CreateTableStmt node
func ParseCreateTableStmt(createStmt string) (*ast.CreateTableStmt, error) {
	createStmt = TrimCharacterSetFromRawCreateTableSQL(createStmt, CharsetWhite()...)
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

func CharsetWhite() []string {
	charsetWhite := strings.Split(os.Getenv(charsetWhiteEnv), ",")
	if len(charsetWhite) == 0 {
		return defaultCharsetWhite
	}
	return charsetWhite
}
