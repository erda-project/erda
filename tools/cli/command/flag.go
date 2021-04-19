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

package command

import (
	"fmt"
	"strconv"
	"strings"
)

type Flag interface {
	Flag()
	DefaultV() string
}

type IntFlag struct {
	Short        string
	Name         string
	Doc          string
	DefaultValue int
}
type FloatFlag struct {
	Short        string
	Name         string
	Doc          string
	DefaultValue float64
}

type BoolFlag struct {
	Short        string
	Name         string
	Doc          string
	DefaultValue bool
}
type StringFlag struct {
	Short        string
	Name         string
	Doc          string
	DefaultValue string
}

type IPFlag struct {
	Short        string
	Name         string
	Doc          string
	DefaultValue string
}

type StringListFlag struct {
	Short        string
	Name         string
	Doc          string
	DefaultValue []string
}

func (IntFlag) Flag()        {}
func (FloatFlag) Flag()      {}
func (BoolFlag) Flag()       {}
func (StringFlag) Flag()     {}
func (IPFlag) Flag()         {}
func (StringListFlag) Flag() {}

func (v IntFlag) DefaultV() string {
	return strconv.Itoa(v.DefaultValue)
}
func (v FloatFlag) DefaultV() string {
	return fmt.Sprintf("%g", v.DefaultValue)
}
func (v BoolFlag) DefaultV() string {
	return fmt.Sprintf("%v", v.DefaultValue)
}
func (v StringFlag) DefaultV() string {
	return v.DefaultValue
}
func (v IPFlag) DefaultV() string {
	return v.DefaultValue
}

// []string{`a`, `b`}
func (v StringListFlag) DefaultV() string {
	if len(v.DefaultValue) == 0 {
		return "nil"
	}
	tmpl := "[]string{%s}"
	elems := []string{}
	for _, e := range v.DefaultValue {
		elems = append(elems, "`"+e+"`")
	}
	return fmt.Sprintf(tmpl, strings.Join(elems, ", "))
}
