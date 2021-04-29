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

package interceptors

import (
	"net/http"
)

// Interceptor .
type Interceptor struct {
	Order   int
	Wrapper func(h http.HandlerFunc) http.HandlerFunc
}

// Interface .
type Interface interface {
	List() []*Interceptor
}

// Interceptors .
type Interceptors []*Interceptor

func (list Interceptors) Len() int           { return len(list) }
func (list Interceptors) Less(i, j int) bool { return list[i].Order < list[j].Order }
func (list Interceptors) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// Config .
type Config struct {
	Order int `file:"order"`
}
