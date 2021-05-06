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

var OPENAPI_EVENT_DOC = apis.ApiSpec{
	Path:      "/api/openapi-event-doc",
	Method:    "GET",
	Scheme:    "http",
	Custom:    getEventDoc,
	Doc:       "获取 openevent 文档",
	IsOpenAPI: true,
}

var (
	eventDocLock sync.Once
	eventDoc     []byte
)

func getEventDocContent() []byte {
	eventDocLock.Do(func() {
		f, err := os.Open("./events_all.json")
		defer f.Close()
		if err != nil {
			logrus.Errorf("getEventDocContent: %v", err)
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			logrus.Errorf("getEventDocContent: %v", err)
		}
		eventDoc = content
	})
	return eventDoc
}

func getEventDoc(rw http.ResponseWriter, req *http.Request) {
	j := getEventDocContent()
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.WriteHeader(200)
	rw.Write(j)
}
