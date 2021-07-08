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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/generate/apistruct"
	"github.com/erda-project/erda/pkg/swagger/oas2"
	"github.com/erda-project/erda/pkg/swagger/oasconv"
)

var docJson = make(apistruct.JSON)

var compile = regexp.MustCompile("\\{(.+?)\\}")

func generateDoc(onlyOpenapi bool, resultfile string) {
	var (
		apisM     = make(map[string][]*apis.ApiSpec)
		apisNames = make(map[*apis.ApiSpec]string)
	)

	for i := range APIs {
		api := &APIs[i]
		if api.Host == "" {
			continue
		}
		title := strings.Split(api.Host, ".")[0]
		apisM[title] = append(apisM[title], api)
		apisNames[api] = APINames[i]
	}

	for title, apiList := range apisM {
		if err := json.Unmarshal([]byte(`{"swagger": "2.0"}`), &docJson); err != nil {
			panic(err)
		}
		docJson["paths"] = make(apistruct.JSON)
		docJson["definitions"] = make(apistruct.JSON)

		for _, api := range apiList {
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
		filename := filepath.Base(resultfile)
		filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".yml"
		filename = title + "-" + filename
		docf, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		docJsonTxt, err := json.Marshal(docJson)
		if err != nil {
			panic(err)
		}

		v2, err := oas2.LoadFromData(docJsonTxt)
		if err != nil {
			panic(errors.Wrap(err, "failed to LoadFromData"))
		}

		v3, err := oasconv.OAS2ConvTo3(v2)
		if err != nil {
			panic(errors.Wrap(err, "failed to OAS2ConvTo3"))
		}

		if v3.Info == nil {
			v3.Info = &openapi3.Info{}
		}
		v3.Info.Title = title
		v3.Info.Version = "default"

		for path, pathItem := range v3.Paths {
			newpath := strings.ReplaceAll(path, "<", "{")
			newpath = strings.ReplaceAll(newpath, ">", "}")
			delete(v3.Paths, path)
			v3.Paths[newpath] = pathItem

			var existsPathParams = make(map[string]bool, 0)
			for _, p := range pathItem.Parameters {
				if p.Value == nil {
					continue
				}
				if p.Value.In == "path" {
					existsPathParams[p.Value.Name] = true
				}
			}

			pns := compile.FindAllString(newpath, -1)
			for _, pn := range pns {
				name := pn[1 : len(pn)-1]
				if _, ok := existsPathParams[name]; ok {
					continue
				}
				existsPathParams[name] = true
				pathParameter := openapi3.NewPathParameter(name)
				pathItem.Parameters = append(pathItem.Parameters, &openapi3.ParameterRef{Value: pathParameter})
			}
		}

		v3j, err := json.Marshal(v3)
		if err != nil {
			panic(errors.Wrap(err, "failed to Marshal v3"))
		}
		v3y, err := oasconv.JSONToYAML(v3j)
		if err != nil {
			panic(errors.Wrap(err, "failed to JSONToYAML"))
		}
		if _, err := docf.Write(v3y); err != nil {
			panic(err)
		}
		_ = docf.Close()
	}

}
