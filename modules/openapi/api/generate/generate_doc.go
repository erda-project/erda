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
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/pkg/swagger/oas3"
)

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
		var v3 = apis.NewSwagger(title)

		for _, api := range apiList {
			if api.IsValidForOperation() {
				panicError(api.AddOperationTo(v3))
			}
		}

		panicError(writeSwagger(resultfile, title, v3))
	}
}

func writeSwagger(filename, title string, v3 *openapi3.Swagger) error {
	filename = filepath.Base(filename)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".yml"
	filename = title + "-" + filename

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	v3Data, err := oas3.MarshalYaml(v3)
	if err != nil {
		return err
	}

	_, err = f.Write(v3Data)
	return err
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}
