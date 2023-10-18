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

package sheets

import (
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

type Namer interface {
	SheetName() string
}

type Importer interface {
	Namer
	DecodeSheet(data *vars.DataForFulfill, s *excel.Sheet) error
	ImporterBeforeCreateIssues
	ImporterAfterCreateIssues
}

type ImporterBeforeCreateIssues interface {
	BeforeCreateIssues(data *vars.DataForFulfill) error
}

type ImporterCreateIssues interface {
	CreateIssues(data *vars.DataForFulfill) error
}

type ImporterAfterCreateIssues interface {
	AfterCreateIssues(data *vars.DataForFulfill) error
}

type Exporter interface {
	Namer
	ExportSheet(data *vars.DataForFulfill) (*RowsForExport, error)
}

type RowsForExport struct {
	Rows          excel.Rows
	SheetHandlers []excel.SheetHandler
}

func NewRowsForExport(rows excel.Rows, sheetHandlers ...excel.SheetHandler) *RowsForExport {
	return &RowsForExport{Rows: rows, SheetHandlers: sheetHandlers}
}

type DefaultImporter struct{}

func (d *DefaultImporter) BeforeCreateIssues(data *vars.DataForFulfill) error { return nil }
func (d *DefaultImporter) AfterCreateIssues(data *vars.DataForFulfill) error  { return nil }
