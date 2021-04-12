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

package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/erda-project/erda/modules/openapi/api/generate/apistruct"
)

var docJson = make(apistruct.JSON)

func generateDoc(onlyOpenapi bool, resultfile string) {
	if err := json.Unmarshal([]byte(`{"swagger": "2.0"}`), &docJson); err != nil {
		panic(err)
	}
	docJson["paths"] = make(apistruct.JSON)
	docJson["definitions"] = make(apistruct.JSON)
	for _, api := range APIs {
		if api.Method == "" {
			continue
		}
		if !api.IsOpenAPI && onlyOpenapi {
			continue
		}
		summary := ""
		doc := strings.Split(strings.TrimSpace(api.Doc), "\n")
		if len(doc) != 0 {
			summary = doc[0]
		}
		var req, resp interface{}
		if api.RequestType == nil {
			req = struct{}{}
		} else {
			req = api.RequestType
		}
		if api.ResponseType == nil {
			resp = struct{}{}
		} else {
			resp = api.ResponseType
		}
		//  如果Group为空，则默认为Path的第二部分
		group := api.Group
		if group == "" {
			pathn := strings.Split(api.Path, "/")
			if len(pathn) >= 1 && pathn[0] == "" {
				pathn = pathn[1:]
			}
			if len(pathn) >= 2 {
				group = pathn[1]
			}
		}

		apistruct.ToJson(api.Path, strings.ToLower(api.Method), summary, group, docJson, req, resp)

	}
	docf, err := os.OpenFile(resultfile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	docJsonTxt, err := json.Marshal(docJson)
	if err != nil {
		panic(err)
	}
	docf.Write(docJsonTxt)

}
