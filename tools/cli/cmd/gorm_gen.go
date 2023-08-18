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
	_ Retriever = &retriever{}
	_ Pager   = &pager{}
)

// CRUDer represents an interface that can perform CRUD operations.
type CRUDer interface {
	// Creator returns a Creator interface that you can create record with.
	Creator(db *gorm.DB) Creator
	// Retriever returns a Retriever that you can select one record with.
	Retriever(db *gorm.DB) Retriever
	// Updater returns an Updater that you can update records with.
	Updater(db *gorm.DB) Updater
	// Deleter returns a Deleter that you can delete records with. 
	Deleter(db *gorm.DB) Deleter
}

// Lister represents an interface that can query multi records with.
type Lister interface {
	// Pager returns a Pager that you can query records by paging with.
	Pager(db*gorm.DB) Pager
}

// Field contains some operations on the model field.
// The reason for generating such an interface is to make it easier to express field operations 
// without having to pass a string when using the gorm interface.
type Field interface {
	// IsNull returns the condition "xx IS NULL"
	IsNull() Where
	// IsNotNull returns the condition "xx IS NOT NULL"
	IsNotNull() Where
	// Equal returns the condition expression "xx = ?, v"
	Equal(v any) Where
	// NotEqual returns the condition expression "xx != ?, v"
	NotEqual(v any) Where
	// In returns the condition expression "xx IN (?), v"
	In(v []any) Where
	// NotIn returns the condition expression "xx NOT IN (?), v" 
	NotIn(v []any) Where
	// LessThan returns the condition expression "xx < ?, v"
	LessThan(v any) Where
	// MoreThan returns the condition expression "xx > ?, v" 
	MoreThan(v any) Where
	// LessEqualThan returns the condition expression "xx <= ?, v" 
	LessEqualThan(v any) Where
	// MoreEqualThan returns the condition expression "xx >= ?, v" 
	MoreEqualThan(v any) Where
	// Set returns a Setter for the Field which will be set in the UPDATE clause. 
	Set(v any) Setter
	// DESC means that the query results are sorted in descending order by the Field.
	// It's used in the last parameter of the method Pager.Paging.
	DESC() string
	// ASC means the query results are in ascending order by the Field.
	// It is used in the last parameter of the method Pager.Paging
	ASC() string
}

// Where denotes a query condition.
// It represents a WHERE clause.
type Where interface {
	Query() any
	Args() []any
}

// Setter indicates that the "UPDATE ... SET ..." statement.
type Setter interface {
	Key() string
	Value() any
}

// Creator is used to create a record.
// It is used to represent an "INSERT INTO" statement.
type Creator interface {
	// Create executes "INSERT INTO" clause on the table.
	// It create a record for the model.
	// It reflects the creation result back to the given model structure pointer.
	Create() error
}

// Deleter is used to delete records.
// It represents a DELETE statement.
type Deleter interface {
	// Where returns a Deleter with the conditions.
	Where(wheres ...Where) Deleter
	// Delete executes "DELETE" clause on the table.
	// It deletes records with the conditions on the table.
	// It returns affected rows or an error.
	Delete() (affects int64, err error)
}

// Updater is used to update a record.
// It represents an UPDATE SET statement.
type Updater interface {
	// Where returns an Updater with the conditions.
	Where(wheres ...Where) Updater
	// Set returns an Updater with fields that will be set in the UPDATE clause.
	Set(setters ...Setter) Updater
	// Update executes "UPDATE ... SET ..." clause.
	// It updates the whole model.
	// It returns affected rows or an error.
    Update() (affects int64, err error)
	// Updates is a short cut for .Set(...).Updates() and updates fields by given.
	Updates(setters ...Setter) (affects int64, err error)
}

// Retriever is used to query for a record that matches a condition.
// It represents the "SELECT ... LIMIT 1" statement.
type Retriever interface {
	// Where returns an Retriever with the conditions.
	Where(wheres ...Where) Retriever
	// Get executes the "SELECT . LIMIT 1" statement.
	// If the query fails, return (false, err).
	// If the query row is not found, return (false, nil), 
	// so the caller doesn't have to decide the err for itself by errors.Is(err, gorm.ErrRecordNotFound).
	// If there is at least one record that matches the condition, return (true, nil), 
	// and reflect the result back to the given model struct pointer.
	Get() (ok bool, err error)
}

// Pager is used to query for multiple matching records.
// It means "select ... offset ... limit ..." statement.
type Pager interface {
	// Where sets conditions for paging.
	Where(wheres ...Where) Pager

	// Paging reflects the results of a paging query to the specified slice, returning count and error.
	// When paging is not needed, pass -1, -1 for size and num.
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

type retriever struct {
	model any
	db    *gorm.DB
	where []Where
}

func (g *retriever) Where(wheres ...Where) Retriever {
	g.where = append(g.where, wheres...)
	return g
}

func (g *retriever) Get() (bool, error) {
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
	if err := db.Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}
	if num > 0 && size >= 0 {
		db = db.Limit(size).Offset((num - 1) * size)
	}
	for _, order := range orders {
		db = db.Order(order)
	}
	if err := db.Find(p.list).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
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

var (
	_ CRUDer = (*{{$.SName}})(nil)
	_ Lister = (*{{$.SName}}List)(nil)
)

{{range $field := .Fields}}
// Field{{.Name}} returns the Field interface{} for the field {{$.TableName}}.{{.Column}}
func (this *{{$.SName}}) Field{{.Name}} () Field { return field{name: "{{.Column}}"} }

{{end}}

// Creator returns a Creator interface that you can create record with.
func (this *{{$.SName}}) Creator(db *gorm.DB) Creator {
	return &creator{db: db, model: this}
}

// Retriever returns a Retriever that you can select one record with.
func (this *{{$.SName}}) Retriever(db *gorm.DB) Retriever {
	return &retriever{
		db:    db.Model(this),
		model: this,
		where: make([]Where, 0),
	}
}

// Updater returns an Updater that you can update records with.
func (this *{{$.SName}}) Updater(db *gorm.DB) Updater {
	return &updater{
		db:      db.Model(this),
		model:   this,
		where:   nil,
		updates: make(map[string]any),
	}
}

// Deleter returns a Deleter that you can delete records with.
func (this *{{$.SName}}) Deleter(db *gorm.DB) Deleter {
	return &deleter{
		db:    db.Model(this),
		model: this,
		where: make([]Where, 0),
	}
}

{{range $field := .Fields}}
// Field{{.Name}} returns the Field interface{} for the field {{$.TableName}}.{{.Column}}
func (list {{$.SName}}List) Field{{.Name}} () Field { return field{name: "{{.Column}}"} }

{{end}}

// Pager returns a Pager that you can query records by paging with.
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
	Example:    "erda-cli gorm gen -f some-file.sql --output app/models",
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
		Source:  "",
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

	// regenerate package common
	(&packageCommon).Reset()
	pkg.Source = fmt.Sprintf("bin/erda-cli gorm gen -f %s -o %s", filename, output)
	if err := template.Must(template.New("package_common").Parse(templatePackage)).Execute(&packageCommon, pkg); err != nil {
		return err
	}

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
