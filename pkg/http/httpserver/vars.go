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

package httpserver

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	infrahttpserver "github.com/erda-project/erda-infra/providers/httpserver"
)

var infrahttpserverVars = infrahttpserver.Vars

func getVars(r *http.Request) (vars map[string]string) {
	// defer decode vars
	defer func() {
		for k, v := range vars {
			decodedVar, err := url.QueryUnescape(v)
			if err != nil {
				continue
			}
			vars[k] = decodedVar
		}
	}()
	// get from mux firstly
	vars = mux.Vars(r)
	if len(vars) > 0 {
		return
	}
	// try get from infrahttpserver when legacyhttpserver only be used as Router, so no vars inject logic executed.
	vars = infrahttpserverVars(r)
	return
}
