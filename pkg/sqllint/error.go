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

package sqllint

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/color"
)

type LintError struct {
	ScriptName string // 脚本名称
	Stmt       string // SQL 语句
	Lint       string // lint 提示
	Line       string // lint 提示所在的行内容
	LintNo     int    // lint 提示所在行行号
}

func NewLintError(script Script, stmt string, lint string, getLine func(line []byte) bool) LintError {
	line, num := getLintLine(script.Data(), []byte(stmt), getLine)
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

// 计算 SQL 脚本发生 lint error 行号
func getLintLine(source, scope []byte, goal func(line []byte) bool) (line string, num int) {
	var firstLine []byte
	scanner := bufio.NewScanner(bytes.NewBuffer(scope))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimPrefix(line, []byte(" "))) > 0 {
			firstLine = line
			break
		}
	}

	idx := bytes.Index(source, scope)
	if idx < 0 {
		return "", -1
	}
	var (
		i       = 0
		lineNum = 0
		minNum  = 0
	)
	scanner = bufio.NewScanner(bytes.NewBuffer(source))
	for scanner.Scan() {
		line := scanner.Bytes()
		i += len(line) + 1
		lineNum++
		if i < idx {
			continue
		}
		if bytes.Equal(line, firstLine) {
			minNum = lineNum
		}
		if goal(line) && !bytes.HasPrefix(bytes.TrimSpace(line), []byte("--")) {
			return string(line), lineNum
		}
		// if bytes.Contains(line, []byte(";")) {
		// 	break
		// }
	}
	return string(firstLine), minNum
}
