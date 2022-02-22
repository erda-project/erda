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

package common

import (
	"fmt"

	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	jsoniter "github.com/json-iterator/go"
)

const (
	JSONSerializer = "json"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func JsonSerializeBatch(ods []odata.ObservableData, compatibility bool) ([]byte, error) {
	return SerializeBatch(JSONSerializer, ods, compatibility)
}

func JSONSerializeSingle(od odata.ObservableData, compatibility bool) ([]byte, error) {
	return SerializeSingle(JSONSerializer, od, compatibility)
}

func SerializeBatch(format string, ods []odata.ObservableData, compatibility bool) ([]byte, error) {
	batch := make([]interface{}, len(ods))
	for idx, item := range ods {
		if compatibility {
			batch[idx] = item.SourceCompatibility()
		} else {
			batch[idx] = item
		}
	}
	return doSerialize(format, batch)
}

func SerializeSingle(format string, od odata.ObservableData, compatibility bool) ([]byte, error) {
	var data interface{}
	if compatibility {
		data = od.SourceCompatibility()
	} else {
		data = od.Source()
	}
	return doSerialize(format, data)
}

func doSerialize(format string, data interface{}) ([]byte, error) {
	switch format {
	case JSONSerializer:
		return json.Marshal(data)
	default:
		return nil, fmt.Errorf("invalid format: %q", format)
	}
}
