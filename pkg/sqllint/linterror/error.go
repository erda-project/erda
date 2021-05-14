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

package linterror

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/color"
	"github.com/erda-project/erda/pkg/sqllint/script"
)

type LintError struct {
	ScriptName string // 脚本名称
	Stmt       string // SQL 语句
	Lint       string // lint 提示
	Line       string // lint 提示所在的行内容
	LintNo     int    // lint 提示所在行行号
}

func New(script script.Script, stmt string, lint string, getLine func(line []byte) bool) LintError {
	line, num := CalcLintLine(script.Data(), []byte(stmt), getLine)
	return LintError{
		ScriptName: script.Name(),
		Stmt:       stmt,
		Lint:       lint,
		Line:       line,
		LintNo:     num,
	}
}

func (e LintError) Error() string {
	scanner := bufio.NewScanner(bytes.NewBufferString(strings.TrimLeft(e.Stmt, "\n")))
	buf := bytes.NewBuffer(nil)
	for scanner.Scan() {
		if line := scanner.Bytes(); bytes.Equal(line, []byte(e.Line)) {
			buf.WriteString("\n~~~> ")
			buf.WriteString(strings.TrimLeft(color.Red(e.Line), "\n"))
		} else {
			buf.WriteString("\n|->  ")
			buf.Write(bytes.TrimPrefix(scanner.Bytes(), []byte("\n")))
		}
	}
	buf.WriteString("\n")
	return fmt.Sprintf("%s:%v: %s: %s\n", e.ScriptName, e.LintNo, e.Lint, buf.String())
}

func (e LintError) StmtName() string {
	p := parser.New()
	node, err := p.ParseOneStmt(e.Stmt, "", "")
	if err != nil {
		return ""
	}
	switch node.(type) {
	case *ast.CreateTableStmt:
		n := node.(*ast.CreateTableStmt)
		if n.Table == nil {
			return ""
		}
		return n.Table.Name.String()
	case *ast.AlterTableStmt:
		n := node.(*ast.AlterTableStmt)
		if n.Table == nil {
			return ""
		}
		return n.Table.Name.String()
	}

	return ""
}
