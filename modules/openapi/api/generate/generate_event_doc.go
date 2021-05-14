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

	"github.com/erda-project/erda/modules/openapi/api/generate/apistruct"
)

var eventDocJSON = make(apistruct.JSON)

func generateEventDoc(onlyOpenEvents bool, resultfile string) {
	if err := json.Unmarshal([]byte(`{"swagger": "2.0"}`), &eventDocJSON); err != nil {
		panic(err)
	}
	js := make(apistruct.JSON)
	eventDocJSON["definitions"] = js
	for _, v := range Events {
		eventStruct := v[0]
		doc := v[1].(string)
		apistruct.EventToJson(eventStruct, doc, js)
	}
	docf, err := os.OpenFile(resultfile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	eventDocTxt, err := json.Marshal(eventDocJSON)
	if err != nil {
		panic(err)
	}
	docf.Write(eventDocTxt)
}
