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

var DOC_JSON = apis.ApiSpec{
	Path:   "/api/openapi/swagger.json",
	Method: "GET",
	Scheme: "http",
	Custom: getDocJSON,
	Doc:    `summary: 返回 swagger.json`,
}

var (
	swaggerJSON     []byte
	swaggerJSONLock sync.Once
)

func getSwaggerJSON() []byte {
	swaggerJSONLock.Do(func() {
		f, err := os.Open("./swagger_all.json")
		defer f.Close()
		if err != nil {
			logrus.Errorf("getSwaggerJSON: %v", err)
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			logrus.Errorf("getSwaggerJSON: %v", err)
		}
		swaggerJSON = content
	})
	return swaggerJSON
}
func getDocJSON(rw http.ResponseWriter, req *http.Request) {
	j := getSwaggerJSON()
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.WriteHeader(200)
	rw.Write(j)
}
