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

package swagger

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/swagger/oas2"
	"github.com/erda-project/erda/pkg/swagger/oas3"
	"github.com/erda-project/erda/pkg/swagger/oasconv"
)

func LoadFromData(spec []byte) (*openapi3.Swagger, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal(spec, &m); err != nil {
		return nil, err
	}
	if _, ok := m["swagger"]; ok {
		v2, err := oas2.LoadFromData(spec)
		if err != nil {
			return nil, err
		}
		return oasconv.OAS2ConvTo3(v2)
	}

	return oas3.LoadFromData(spec)
}

// 序列化为 json 格式
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// 序列化为 yaml 格式
func MarshalYAML(v interface{}) ([]byte, error) {
	// 之所以要先序列化成 json, 再转换为 yaml,
	// 是因为 *openapi.Swagger 对象实现了 MarshalJSON,
	// 有特定的序列化规则
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return oasconv.JSONToYAML(data)
}
