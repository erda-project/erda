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

//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"io"
//	"io/ioutil"
//	"os"
//	"path"
//
//	"github.com/bitly/go-simplejson"
//	"github.com/ghodss/yaml"
//)
//
//const (
//	dstActionFile = "../../assets/actions-spec.json"
//)
//
//func main() {
//	var err error
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Printf("failed to generate %s: %v\n", dstActionFile, r)
//			os.Remove(dstActionFile)
//		}
//	}()
//	if err = os.MkdirAll(path.Dir(dstActionFile), 0755); err != nil {
//		panic(err)
//	}
//	w, err := os.OpenFile(dstActionFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
//	if err != nil {
//		panic(err)
//	}
//
//	actionDir := "../../../../cmd/actions"
//
//	fs, err := ioutil.ReadDir(actionDir)
//	if err != nil {
//		panic(err)
//	}
//	actionJson := "["
//	for _, f := range fs {
//		b, err := ioutil.ReadFile(path.Join(actionDir, f.Name(), "spec.yml"))
//		if err != nil {
//			panic(err)
//		}
//		jsonBytes, err := yaml.YAMLToJSON(b)
//		newJson, err := simplejson.NewJson(jsonBytes)
//		if err != nil {
//			panic(err)
//		}
//		sourceJson := newJson.Get("source")
//		if sourceJson != nil {
//			setDefaultArrayValue(newJson, "source")
//		}
//
//		paramsJson := newJson.Get("params")
//		if paramsJson != nil {
//			setDefaultArrayValue(newJson, "params")
//		}
//
//		marshalJSON, _ := newJson.MarshalJSON()
//		actionJson += string(marshalJSON) + ","
//	}
//	if actionJson[len(actionJson)-1] == ',' {
//		actionJson = actionJson[0:len(actionJson)-1] + "]"
//	} else {
//		actionJson += "]"
//	}
//	var prettyJSON bytes.Buffer
//	err = json.Indent(&prettyJSON, []byte(actionJson), "", "\t")
//	if err != nil {
//		panic(err)
//	}
//
//	io.WriteString(w, prettyJSON.String())
//
//	fmt.Println(prettyJSON.String())
//}
//
//var defaultsOptions = map[string]interface{}{
//	"type":     "string",
//	"required": false,
//	"desc":     "",
//}
//
//func setDefaultArrayValue(val *simplejson.Json, key string) {
//	structArray, _ := val.Get(key).Array()
//	for i := range structArray {
//		params := val.Get(key).GetIndex(i)
//		for k, v := range defaultsOptions {
//			_, exist := params.CheckGet(k)
//			if !exist {
//				params.Set(k, v)
//			}
//		}
//
//		tye := params.Get("type").MustString()
//		if tye == "struct" || tye == "struct_array" {
//			setDefaultArrayValue(params, "struct")
//		}
//	}
//}
