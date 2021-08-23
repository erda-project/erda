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
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
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
