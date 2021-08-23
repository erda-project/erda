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

package linters

import (
	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

type DDLDMLLinter struct {
	baseLinter
}

func NewDDLDMLLinter(script script.Script) rules.Rule {
	return &DDLDMLLinter{newBaseLinter(script)}
}

func (l *DDLDMLLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case ast.DDLNode, ast.DMLNode:
	default:
		l.err = linterror.New(l.s, l.text,
			"language type error: only support DDL, DML, not support DCL, TCL",
			func(line []byte) bool {
				return true
			})
	}

	return in, true
}

func (l *DDLDMLLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, false
}

func (l *DDLDMLLinter) Error() error {
	return l.err
}
