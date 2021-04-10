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

package validator

import (
	"context"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
)

// 校验OpenAPI格式的合法性
func ValidateOpenapi(data []byte) (*openapi3.Swagger, error) {
	o, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(data)
	if err != nil {
		return nil, err
	}

	if err := o.Validate(context.Background()); err != nil {
		return nil, err
	}
	if o.Info == nil {
		return nil, errors.New("no Info in the OpenAPI")
	}
	return o, nil
}

// 校验 openapi 3 合法性
func ValidateOAS3(data []byte, protocol string) (oas3 *openapi3.Swagger, err error) {
	switch strings.ToLower(protocol) {
	case OAS3JSON:
	case OAS3YAML:
		if data, err = yaml2.ToJSON(data); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return ValidateOpenapi(data)
}
