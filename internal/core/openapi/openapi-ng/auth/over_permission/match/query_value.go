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
	"net/url"
	"strings"
)

func init() {
	registry("query", queryValue{})
}

type queryValue struct {
}

/*
http://aaaa.com?aaa=1&bbb=2
query:aaa -> aaa:1
query:bbb -> bbb:2
*/
func (q queryValue) get(expr string, r *http.Request) interface{} {
	return find(expr, r.URL.Query())
}

func find(expr string, values url.Values) interface{} {
	path := strings.Split(expr, keyDelim)
	return search(path, values)
}

func search(path []string, values url.Values) interface{} {
	if len(path) == 0 {
		return nil //missing
	}
	// query no have nested!, let`s keep simple
	return values.Get(path[0])
}
