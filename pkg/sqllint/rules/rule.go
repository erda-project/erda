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

package rules

import (
	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/sqllint/script"
)

// Rule is an Error and SQL ast visitor,
// can accept a SQL stmt and lint it.
type Rule interface {
	ast.Visitor

	Error() error
}

// Ruler is a function that returns a Rule interface
type Ruler func(script script.Script) Rule
