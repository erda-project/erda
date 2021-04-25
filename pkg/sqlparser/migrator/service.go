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
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"gorm.io/gorm"
)

// Module is the list of Script
type Module []*Script

// BaselineEqualCloud check baseline script schema is equal with cloud schema or not
func (s Module) BaselineEqualCloud(tx *gorm.DB) *Equal {
	tableNames := s.BaselineTableNames()
	cloud := NewSchema()
	for _, tableName := range tableNames {
		raw := "SHOW CREATE TABLE " + tableName
		this := tx.Raw(raw)
		if err := this.Error; err != nil {
			return &Equal{
				equal:  false,
				reason: fmt.Sprintf("failed to exec %s", raw),
			}
		}
		var _ig, create string
		if err := this.Row().Scan(&_ig, &create); err != nil {
			return &Equal{
				equal:  false,
				reason: fmt.Sprintf("failed to Scan create table stmt, raw: %s", raw),
			}
		}
		node, err := parser.New().ParseOneStmt(create, "", "")
		if err != nil {
			return &Equal{
				equal:  false,
				reason: fmt.Sprintf("failed to ParseOneStmt: %s, raw: %s", create, raw),
			}
		}
		node.Accept(cloud)
	}

	return s.BaselineSchema().Equal(cloud)
}

func (s Module) Schema() *Schema {
	schema := NewSchema()
	for _, script := range s {
		for _, ddl := range script.DDLNodes() {
			ddl.Accept(schema)
		}
	}
	return schema
}

func (s Module) BaselineSchema() *Schema {
	schema := NewSchema()
	for _, script := range s {
		if !script.IsBaseline() {
			continue
		}
		for _, ddl := range script.DDLNodes() {
			ddl.Accept(schema)
		}
	}
	return schema
}

func (s Module) TableNames() []string {
	var names []string
	for _, script := range s {
		for _, ddl := range script.DDLNodes() {
			if create, ok := ddl.(*ast.CreateTableStmt); ok {
				if create.Table != nil {
					names = append(names, create.Table.Name.String())
				}
			}
		}
	}
	return names
}

func (s Module) BaselineTableNames() []string {
	var names []string
	for _, script := range s {
		if !script.IsBaseline() {
			continue
		}

		for _, ddl := range script.DDLNodes() {
			create, ok := ddl.(*ast.CreateTableStmt)
			if !ok || create.Table == nil {
				continue
			}
			names = append(names, create.Table.Name.String())
		}
	}

	return names
}

func (s Module) Sort() {
	sort.Slice(s, func(i, j int) bool {
		return strings.TrimSuffix(s[i].Name, filepath.Ext(s[i].Name)) < strings.TrimSuffix(s[j].Name, filepath.Ext(s[j].Name))
	})
	sort.Slice(s, func(i, j int) bool {
		return s[i].IsBaseline() && !s[j].IsBaseline()
	})
}

func (s Module) Filenames() []string {
	var names []string
	for _, script := range s {
		names = append(names, script.Name)
	}
	return names
}
