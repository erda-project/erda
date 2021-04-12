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

package openapi

import (
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var OPENAPI_DOC = apis.ApiSpec{
	Path:      "/api/openapi-doc",
	Method:    "GET",
	Scheme:    "http",
	Custom:    getOpenAPIDoc,
	Doc:       "返回 openapi 文档",
	IsOpenAPI: true,
}

var (
	openApiDocLock sync.Once
	openApiDoc     []byte
)

func getOpenAPIDocContent() []byte {
	openApiDocLock.Do(func() {
		f, err := os.Open("./swagger.json")
		defer f.Close()
		if err != nil {
			logrus.Errorf("getOpenAPIDoc: %v", err)
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			logrus.Errorf("getOpenAPIDoc: %v", err)
		}
		openApiDoc = content
	})
	return openApiDoc
}

func getOpenAPIDoc(rw http.ResponseWriter, req *http.Request) {
	j := getOpenAPIDocContent()
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.WriteHeader(200)
	rw.Write(j)
}
