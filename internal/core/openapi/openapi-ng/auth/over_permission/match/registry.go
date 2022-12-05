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

package match

import (
	"net/http"
	"strings"
	"sync"
)

var ValueFunc map[string]value
var once sync.Once

const (
	typeDelim = ":"
	keyDelim  = "."
)

type value interface {
	get(expr string, r *http.Request) interface{}
}

func registry(t string, funcs value) {
	once.Do(func() {
		if ValueFunc == nil {
			ValueFunc = make(map[string]value)
		}
	})
	ValueFunc[t] = funcs
}

// Get use expr to match function
func Get(express []string, request *http.Request) map[string]interface{} {
	if len(express) == 0 {
		return nil
	}
	typeDelimPoint := strings.IndexAny(express[0], typeDelim)
	if typeDelimPoint <= 0 || typeDelimPoint == 1 {
		return nil
	}

	m := make(map[string]interface{})
	for _, expr := range express {
		typp := expr[:typeDelimPoint]
		path := expr[typeDelimPoint+1:]
		if len(path) == 0 {
			continue
		}
		if f, ok := ValueFunc[typp]; ok {
			m[path] = f.get(path, request)
		}
	}
	return m
}
