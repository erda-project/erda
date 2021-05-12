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

package linters

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/sqllint/linterror"
	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/pkg/sqllint/script"
)

type IndexNameLinter struct {
	baseLinter
}

func NewIndexNameLinter(script script.Script) rules.Rule {
	return &IndexNameLinter{newBaseLinter(script)}
}

func (l *IndexNameLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	constraint, ok := in.(*ast.Constraint)
	if !ok {
		return in, false
	}

	switch constraint.Tp {
	case ast.ConstraintIndex:
		if !strings.HasPrefix(constraint.Name, "idx_") {
			l.err = linterror.New(l.s, l.text, "index name error: normal index name should start with idx_",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), []byte("index")) &&
						bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(constraint.Name)))
				})
			return in, true
		}
	case ast.ConstraintUniq, ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
		if !strings.HasPrefix(constraint.Name, "uk_") {
			l.err = linterror.New(l.s, l.text, "index name error: unique index name should start with uk_",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), []byte("unique")) &&
						bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(constraint.Name)))
				})
			return in, true
		}
	}

	return in, true
}

func (l *IndexNameLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IndexNameLinter) Error() error {
	return l.err
}
