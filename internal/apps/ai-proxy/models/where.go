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

package models

type WhereField interface {
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
}

type whereField struct {
	fieldName string
}

func (w whereField) IsNull() Where {
	return where{query: w.fieldName + " is null"}
}

func (w whereField) IsNotNull() Where {
	return where{query: w.fieldName + " is not null"}
}

func (w whereField) Equal(v any) Where {
	return where{query: w.fieldName + " == ?", args: []any{v}}
}

func (w whereField) NotEqual(v any) Where {
	return where{query: w.fieldName + " != ?", args: []any{v}}
}

func (w whereField) In(v []any) Where {
	return where{query: w.fieldName + " in ?", args: []any{v}}
}

func (w whereField) NotIn(v []any) Where {
	return where{query: w.fieldName + " not in ?", args: []any{v}}
}

func (w whereField) LessThan(v any) Where {
	return where{query: w.fieldName + " < ?", args: []any{v}}
}

func (w whereField) MoreThan(v any) Where {
	return where{query: w.fieldName + " > ?", args: []any{v}}
}

func (w whereField) LessEqualThan(v any) Where {
	return where{query: w.fieldName + " <= ?", args: []any{v}}
}

func (w whereField) MoreEqualThan(v any) Where {
	return where{query: w.fieldName + " >= ?", args: []any{v}}
}

var _ WhereField = whereField{}

type Source struct{}

type Where interface {
	Query() any
	Args() []any
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
