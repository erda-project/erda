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

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	OAS2JSON = "oas2-json"
	OAS2YAML = "oas2-yaml"
	OAS3YAML = "oas3-yaml"
	OAS3JSON = "oas3-json"
)

// 校验 swagger 2.x 格式合法性
func ValidateOAS2(data []byte, protocol string) (swagger *openapi3.Swagger, err error) {
	switch strings.ToLower(protocol) {
	case OAS2JSON:
	case OAS2YAML:
		if data, err = yaml2.ToJSON(data); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	v2Swagger := new(openapi2.Swagger)
	if err := v2Swagger.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	v3Swagger, err := openapi2conv.ToV3Swagger(v2Swagger)
	if err != nil {
		return nil, err
	}

	if err = v3Swagger.Validate(context.Background()); err != nil {
		return nil, err
	}
	if v3Swagger.Info == nil {
		return nil, errors.New("no Info in the swagger")
	}
	return v3Swagger, nil
}

// 对 swagger 合法性进行基础性校验
func BasicallyValidateSwagger(data []byte) (*spec.Swagger, error) {
	doc, err := loads.Analyzed(data, "")
	if err != nil {
		return nil, err
	}
	s := doc.Spec()
	if s == nil {
		return nil, errors.New("swagger is invalid")
	}
	if s.Info == nil {
		return nil, errors.New("swagger is invalid: no info")
	}
	return s, nil
}
