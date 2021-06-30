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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/database/pyorm/pattern"
	"github.com/erda-project/erda/pkg/database/sqlparser/ddlreverser"
)

const (
	ScriptTypeSQL    ScriptType = ".sql"
	ScriptTypePython ScriptType = ".py"
)

type ScriptType string

const (
	baseScriptLabel  = "# MIGRATION_BASE"
	baseScriptLabel2 = "-- MIGRATION_BASE"
	baseScriptLabel3 = "/* MIGRATION_BASE */"
)

// Script is the object from a SQL script file, it contains many statements.
type Script struct {
	// path from repo root, filepath.Base(script.Name) for base filename
	Name      string
	Rawtext   []byte
	Reversing []string
	Nodes     []ast.StmtNode
	Pending   bool
	Record    *HistoryModel
	Workdir   string
	Type      ScriptType
	isBase    bool
}

// NewScript read the local file, parse data as SQL AST nodes, and mark script IsBase or not
func NewScript(workdir, pathFromRepoRoot string) (*Script, error) {
	data, err := ioutil.ReadFile(filepath.Join(workdir, pathFromRepoRoot))
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadFile")
	}
	data = bytes.TrimLeftFunc(data, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})

	var (
		s = &Script{
			Name:      pathFromRepoRoot,
			Rawtext:   data,
			Reversing: nil,
			Nodes:     nil,
			Pending:   true,
			Record:    nil,
			Workdir:   workdir,
			Type:      "",
			isBase: bytes.HasPrefix(data, []byte(baseScriptLabel)) ||
				bytes.HasPrefix(data, []byte(baseScriptLabel2)) ||
				bytes.HasPrefix(data, []byte(baseScriptLabel3)),
		}
	warns []error
	)
	switch ext := filepath.Ext(s.Name); {
	case strings.EqualFold(ext, string(ScriptTypeSQL)):
		s.Type = ScriptTypeSQL
		s.Nodes, warns, err = parser.New().Parse(string(data), "", "")
		if err != nil {
			return nil, errors.Wrap(err, "failed to Parse file text data")
		}
		for _, warn := range warns {
			logrus.Fatalln(warn)
		}
	case strings.EqualFold(ext, string(ScriptTypePython)):
		s.Type = ScriptTypePython
	default:
		return nil, errors.Errorf("invalid script type: only support .sql or .py file: %s", ext)
	}

	for _, node := range s.Nodes {
		// ignore C-Style comments
		text := strings.TrimLeftFunc(node.Text(), func(r rune) bool {
			return r == ' ' || r == '\n' || r == '\t' || r == '\r'
		})
		if strings.Contains(text, "/*!") && strings.Contains(text, "*/") {
			continue
		}
		if strings.HasPrefix(text, "/*!") {
			continue
		}
		node.SetText(text)

		switch node.(type) {
		case ast.DDLNode, ast.DMLNode, *ast.SetStmt:
			s.Nodes = append(s.Nodes, node)
		default:
			if s.IsBaseline() {
				continue
			}
			return nil, errors.Errorf("only support DDL and DML, SQL: %s", node.Text())
		}
	}

	return s, nil
}

func (s *Script) DDLNodes() []ast.DDLNode {
	var results []ast.DDLNode
	for _, node := range s.Nodes {
		if strings.HasPrefix(strings.TrimPrefix(node.Text(), " "), "/*!") {
			continue
		}

		switch node.(type) {
		// ignore LockTablesStmt or UnlockTablesStmt
		case *ast.LockTablesStmt, *ast.UnlockTablesStmt:
		case ast.DDLNode:
			results = append(results, node.(ast.DDLNode))
		}
	}
	return results
}

func (s *Script) DMLNodes() []ast.StmtNode {
	var results []ast.StmtNode
	for _, node := range s.Nodes {
		if strings.HasPrefix(strings.TrimPrefix(node.Text(), " "), "/*!") {
			continue
		}

		switch node.(type) {
		// note: process SetStmt as DML
		case *ast.SetStmt, ast.DMLNode:
			results = append(results, node)
		}
	}
	return results
}

func (s *Script) Checksum() string {
	hash := sha256.New()
	hash.Write(s.Rawtext)
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Script) IsBaseline() bool {
	return s.isBase
}

// Install installs the script in database
func (s *Script) Install(dsn string, begin func() *gorm.DB, after func(tx *gorm.DB, err error)) (err error) {
	if s.Type == ScriptTypeSQL {
		return s.installSQL(begin, after)
	}

	return s.installPy(dsn)
}

func (s *Script) installSQL(begin func() *gorm.DB, after func(tx *gorm.DB, err error)) (err error) {
	tx := begin()
	defer after(tx, err)

	s.Reversing = nil

	for _, node := range s.DDLNodes() {
		var (
			reverse string
			ok      bool
		)
		reverse, ok, err = ddlreverser.ReverseDDLWithSnapshot(DB(), node)
		if err != nil {
			return errors.Wrapf(err, "failed to generate reversed DDL. "+
				"Ther script name: %s, the SQL: %s", s.Name, node.Text())
		}
		if ok {
			s.Reversing = append(s.Reversing, reverse)
		}

		if err = Exec(node.Text()).Error; err != nil {
			return errors.Wrapf(err, "failed to pre-migrate schema SQL, all migrations will be rolled back. "+
				"The script name: %s, the SQL: %s", s.Name, node.Text())
		}
	}

	for _, node := range s.DMLNodes() {
		if err = tx.Exec(node.Text()).Error; err != nil {
			return errors.Wrapf(err, "failed to pre-migrate data SQL, all migration will be rolled back. "+
				"The script filename: %s, the data SQL:\n%s", s.Name, node.Text())
		}
	}

	return nil
}

func (s *Script) installPy(dsn string) (err error) {
	//"root:12345678@(localhost:3306)/"
	dsnConfig, err := mysql.ParseDSN(dsn)
	if err != nil {
		return errors.Wrap(err, "failed to ParseDSN")
	}

	settings := pattern.Settings{
		Engine:        pattern.DjangoMySQLEngine,
		User:          dsnConfig.User,
		Password:      dsnConfig.Passwd,
		Host:          dsnConfig.Addr,
		Port:          0,
		Name:          dsnConfig.DBName,
		TimeZone:      dsnConfig.Loc.String(),
		InstalledApps: strings.TrimSuffix(filepath.Base(s.Name), filepath.Ext(s.Name)),
	}

	var buf = bytes.NewBuffer(nil)
	if err = pattern.GenSettings(buf, settings); err != nil {
		return errors.Wrap(err, "failed to GenSettings")
	}
	buf.WriteString("\n")
	buf.Write(s.Rawtext)

	// mkdir

	// write buf to file

	// write entrypoint file

	// run python

	// rm dir

	return errors.New("not implement")
}