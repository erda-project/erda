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

package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
)

const license = `// Copyright (c) 2021 Terminus, Inc.
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
// limitations under the License.`

const crudFile = `
import (
	"errors"
    "reflect"

	"gorm.io/gorm"
)

var (
	_ Field   = field{}
	_ Where   = where{}
	_ Setter  = setter{}
	_ Creator = &creator{}
	_ Deleter = &deleter{}
	_ Updater = &updater{}
	_ Getter  = &getter{}
	_ Pager   = &pager{}
)

type Field interface {
	IsNull() Where
	IsNotNull() Where
	Equal(v any) Where
	NotEqual(v any) Where
	In([]any) Where
	NotIn([]any) Where
	LessThan(v any) Where
	MoreThan(v any) Where
	LessEqualThan(v any) Where
	MoreEqualThan(v any) Where
	Set(v any) Setter
	DESC() string
	ASC() string
}

type Where interface {
	Query() any
	Args() []any
}

type Setter interface {
	Key() string
	Value() any
}
type Creator interface {
	Create() error
}

type Deleter interface {
	Where(wheres ...Where) Deleter
	Delete() (affects int64, err error)
}

type Updater interface {
	Where(wheres ...Where) Updater
	Set(setters ...Setter) Updater
	Updates(setters ...Setter) (affects int64, err error)
    Update() (affects int64, err error)
}

type Getter interface {
	Where(wheres ...Where) Getter
	Get() (ok bool, err error)
}

type Pager interface {
	Where(wheres ...Where) Pager
	Paging(size, num int, orders ...string) (total int64, err error)
}
type field struct {
	name string
}

func (w field) IsNull() Where {
	return where{query: w.name + " is null"}
}

func (w field) IsNotNull() Where {
	return where{query: w.name + " is not null"}
}

func (w field) Equal(v any) Where {
	return where{query: w.name + " = ?", args: []any{v}}
}

func (w field) NotEqual(v any) Where {
	return where{query: w.name + " != ?", args: []any{v}}
}

func (w field) In(v []any) Where {
	return where{query: w.name + " in ?", args: []any{v}}
}

func (w field) NotIn(v []any) Where {
	return where{query: w.name + " not in ?", args: []any{v}}
}

func (w field) LessThan(v any) Where {
	return where{query: w.name + " < ?", args: []any{v}}
}

func (w field) MoreThan(v any) Where {
	return where{query: w.name + " > ?", args: []any{v}}
}

func (w field) LessEqualThan(v any) Where {
	return where{query: w.name + " <= ?", args: []any{v}}
}

func (w field) MoreEqualThan(v any) Where {
	return where{query: w.name + " >= ?", args: []any{v}}
}

func (w field) DESC() string {
	return w.name + " DESC"
}

func (w field) ASC() string {
	return w.name + " ASC"
}

func (w field) Set(v any) Setter {
	return &setter{key: w.name, value: v}
}

type where struct {
	query any
	args  []any
}

func (w where) Query() any {
	return w.query
}

func (w where) Args() []any {
	return w.args
}
type setter struct {
	key   string
	value any
}

func (s setter) Key() string {
	return s.key
}

func (s setter) Value() any {
	return s.value
}
type creator struct {
	db    *gorm.DB
	model any
}

func (c *creator) Create() error {
	return c.db.Create(c.model).Error
}

type deleter struct {
	db    *gorm.DB
	model any
	where []Where
}

func (d *deleter) Where(where ...Where) Deleter {
	d.where = append(d.where, where...)
	return d
}

func (d *deleter) Delete() (int64, error) {
	var db = d.db
	for _, w := range d.where {
		db = db.Where(w.Query(), w.Args()...)
	}
	err := db.Delete(d.model).Error
	return db.RowsAffected, err
}

type updater struct {
	db      *gorm.DB
    model   interface{ TableName() string }
	where   []Where
	updates map[string]any
}

func (u *updater) Where(where ...Where) Updater {
	u.where = append(u.where, where...)
	return u
}

func (u *updater) Set(set ...Setter) Updater {
	for _, item := range set {
		u.updates[item.Key()] = item.Value()
	}
	return u
}

func (u *updater) Updates(set ...Setter) (int64, error) {
	if len(set) > 0 {
		return u.Set(set...).Updates()
	}

	var db = u.db
	for _, w := range u.where {
		db = db.Where(w.Query(), w.Args()...)
	}
	err := db.Updates(u.updates).Error
	return db.RowsAffected, err
}

func (u *updater) Update() (int64, error) {
	var db = u.db
	for _, w := range u.where {
		db = db.Where(w.Query(), w.Args()...)
	}
	err := db.Updates(u.model).Error
	return db.RowsAffected, err
}

type getter struct {
	model any
	db    *gorm.DB
	where []Where
}

func (g *getter) Where(wheres ...Where) Getter {
	g.where = append(g.where, wheres...)
	return g
}

func (g *getter) Get() (bool, error) {
	var db = g.db
	for _, w := range g.where {
		db = db.Where(w.Query(), w.Args()...)
	}
	err := db.First(g.model).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

type pager struct {
	db    *gorm.DB
	list  any
	where []Where
}

func (p *pager) Where(wheres ...Where) Pager {
	p.where = append(p.where, wheres...)
	return p
}

func (p *pager) Paging(size, num int, orders ...string) (int64, error) {
	var db = p.db.Model(reflect.New(reflect.TypeOf(p.list).Elem().Elem()).Interface())
	for _, w := range p.where {
		db = db.Where(w.Query(), w.Args()...)
	}
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if num > 0 && size >= 0 {
		db = db.Limit(size).Offset((num - 1) * size)
	}
	for _, order := range orders {
		db = db.Order(order)
	}
	return count, db.Find(p.list).Error
}
`

const templatePackage = `{{.License}}

// Code generated by erda-cli. DO NOT EDIT.
// Source: {{.Source}}

package {{.Package}}
`
const templateModel = `
import (
    "time"

    "github.com/erda-project/erda-infra/providers/mysql/v2/plugins/fields"
)

// {{.SName}} is the table {{.TableName}}
type {{.SName}} struct { {{range .Fields}}
    {{.Name}} {{.Type}} {{.Tag}}{{end}}
}

// TableName returns the table name {{.TableName}}
func (*{{.SName}}) TableName() string { return "{{.TableName}}" }

type {{.SName}}List []*{{.SName}}
`

const templateCrud = `
import (
	"gorm.io/gorm"
)

{{range $field := .Fields}}
// Field{{.Name}} returns the Field interface{} for the field {{$.TableName}}.{{.Column}}
func (this *{{$.SName}}) Field{{.Name}} () Field { return field{name: "{{.Column}}"} }

{{end}}

func (this *{{$.SName}}) Creator(db *gorm.DB) Creator {
	return &creator{db: db, model: this}
}

func (this *{{$.SName}}) Deleter(db *gorm.DB) Deleter {
	return &deleter{
		db:    db.Model(this),
		model: this,
		where: make([]Where, 0),
	}
}

func (this *{{$.SName}}) Updater(db *gorm.DB) Updater {
	return &updater{
		db:      db.Model(this),
		model:   this,
		where:   nil,
		updates: make(map[string]any),
	}
}

func (this *{{$.SName}}) Getter(db *gorm.DB) Getter {
	return &getter{
		db:    db.Model(this),
		model: this,
		where: make([]Where, 0),
	}
}

{{range $field := .Fields}}
// Field{{.Name}} returns the Field interface{} for the field {{$.TableName}}.{{.Column}}
func (list {{$.SName}}List) Field{{.Name}} () Field { return field{name: "{{.Column}}"} }

{{end}}

func (list *{{$.SName}}List) Pager(db *gorm.DB) Pager {
	return &pager{
		db:    db,
		list:  list,
		where: nil,
	}
}
`

var upperWords = map[string]struct{}{
	"id":     {},
	"http":   {},
	"ai":     {},
	"api":    {},
	"sha256": {},
}

var GormGen = command.Command{
	ParentName: "Gorm",
	Name:       "gen",
	ShortHelp:  "gen Go struct from create table stmt",
	LongHelp:   "gen Go struct from create table stmt",
	Example:    "erda-cli gorm gen --create-table 'create table t (id varchar(64))' --output app/models",
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "f",
			Name:         "filename",
			Doc:          Doc("the create table stmt", "创建表的 SQL 文件", Required),
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "o",
			Name:         "output",
			Doc:          Doc("the output directory", "输出的目录", Required),
			DefaultValue: "",
		},
	},
	Run: RunGormGen,
}

type Package struct {
	License string `json:"license"`
	Source  string
	Package string
}

type Table struct {
	SName     string
	TableName string
	Fields    []Field
}

type Field struct {
	Name   string
	Type   string
	Tag    string
	Column string
}

func RunGormGen(ctx *command.Context, filename, output string) error {
	ctx.Info("RunGormGen, filename: %s, output: %s",
		filename, output)

	// generate package common
	var packageCommon bytes.Buffer
	var pkg = Package{
		License: license,
		Source:  filename,
		Package: path.Base(output),
	}
	if err := template.Must(template.New("package_common").Parse(templatePackage)).Execute(&packageCommon, pkg); err != nil {
		return err
	}

	var files = make(map[string][]byte)

	// generate curd.go
	crudGo := packageCommon.String() + crudFile
	source, err := format.Source([]byte(crudGo))
	if err != nil {
		return errors.Wrap(err, "failed to format.Source where.go")
	}
	files[path.Join(output, "crud.go")] = source

	// generate every table
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	nodes, _, err := parser.New().Parse(string(data), "", "")
	if err != nil {
		return err
	}
	for _, node := range nodes {
		stmt, ok := node.(*ast.CreateTableStmt)
		if !ok {
			continue
		}
		table := Stmt2Struct(stmt)
		// generate <model>.go
		var tableBuf bytes.Buffer
		if err := template.Must(template.New(table.SName).Parse(templateModel)).Execute(&tableBuf, table); err != nil {
			return err
		}
		source, err := format.Source(append(packageCommon.Bytes(), tableBuf.Bytes()...))
		if err != nil {
			return errors.Wrap(err, "failed to format.Source "+table.TableName+".go")
		}
		files[path.Join(output, table.TableName+".go")] = source

		// generate <model>_where.go
		var whereBuf bytes.Buffer
		if err := template.Must(template.New(table.SName+"Where").Parse(templateCrud)).Execute(&whereBuf, table); err != nil {
			return err
		}
		source, err = format.Source(append(packageCommon.Bytes(), whereBuf.Bytes()...))
		if err != nil {
			return errors.Wrap(err, "failed to format.Source "+table.TableName+"_crud.go")
		}
		files[path.Join(output, table.TableName+"_curd.go")] = source
	}

	for k, v := range files {
		if err := os.WriteFile(k, v, 0644); err != nil {
			return errors.Wrapf(err, "failed to write file %s", k)
		}
	}

	return nil
}

func Stmt2Struct(stmt *ast.CreateTableStmt) *Table {
	var t = Table{
		SName:     TableNameToStructName(stmt.Table.Name.String()),
		TableName: stmt.Table.Name.String(),
		Fields:    nil,
	}
	for _, col := range stmt.Cols {
		name := TableNameToStructName(col.Name.String())
		tag := strings.ToLower(name[:1]) + name[1:]
		tag = strconv.Quote(tag)
		t.Fields = append(t.Fields, Field{
			Name:   name,
			Type:   SQLTypeToGoType(name, col.Tp.Tp),
			Tag:    fmt.Sprintf("`gorm:\"column:%s;type:%s\" json:%s yaml:%s`", col.Name.String(), col.Tp.String(), tag, tag),
			Column: col.Name.String(),
		})
	}
	return &t
}

func TableNameToStructName(name string) string {
	var result string
	words := strings.Split(name, "_")
	for _, word := range words {
		if _, ok := upperWords[strings.ToLower(word)]; ok {
			result += strings.ToUpper(word)
			continue
		}
		result += strings.Title(word)
	}
	return result
}

func SQLTypeToGoType(name string, tp byte) string {
	switch tp {
	case mysql.TypeDecimal, mysql.TypeNewDecimal:
		return "string"
	case mysql.TypeTiny, mysql.TypeBit:
		return "bool"
	case mysql.TypeShort, mysql.TypeYear:
		return "int"
	case mysql.TypeLong, mysql.TypeLonglong:
		return "int64"
	case mysql.TypeFloat, mysql.TypeDouble, mysql.TypeInt24:
		return "float64"
	case mysql.TypeTimestamp, mysql.TypeDatetime, mysql.TypeDate, mysql.TypeNewDate, mysql.TypeDuration:
		if strings.EqualFold(name, "deletedAt") {
			return "fields.DeletedAt"
		}
		return "time.Time"
	case mysql.TypeVarchar, mysql.TypeEnum, mysql.TypeSet, mysql.TypeJSON,
		mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob,
		mysql.TypeVarString, mysql.TypeString:
		if strings.EqualFold(name, "id") {
			return "fields.UUID"
		}
		return "string"
	default:
		return "string"
	}
}
