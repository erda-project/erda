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

package oasconv

import (
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

const (
	OAS2JSON Protocol = "oas2-json"
	OAS2YAML Protocol = "oas2-yaml"
	OAS3YAML Protocol = "oas3-yaml"
	OAS3JSON Protocol = "oas3-json"
)

type Protocol string

func (protocol Protocol) String() string {
	return string(protocol)
}

func OAS3ConvTo2(v3 *openapi3.Swagger) (v2 *openapi2.Swagger, err error) {
	if v3 == nil {
		return nil, errors.New("swagger is nil")
	}

	return openapi2conv.FromV3Swagger(v3)
}

func OAS2ConvTo3(v2 *openapi2.Swagger) (v3 *openapi3.Swagger, err error) {
	if v2 == nil {
		return nil, errors.New("swagger is nil")
	}

	v3, err = openapi2conv.ToV3Swagger(v2)
	if err != nil {
		return nil, err
	}

	if v2.Host != "" {
		v3.Servers = append(v3.Servers, &openapi3.Server{
			URL:         v2.Host,
			Description: "",
			Variables:   nil,
		})
	}

	for _, scheme := range v2.Schemes {
		v3.Servers = append(v3.Servers, &openapi3.Server{
			URL:         scheme,
			Description: "",
			Variables:   nil,
		})
	}

	if strings.HasPrefix(v2.BasePath, "/") {
		paths := make(openapi3.Paths, len(v3.Paths))
		for k := range v3.Paths {
			paths[filepath.Join(v2.BasePath, k)] = v3.Paths[k]
		}
		v3.Paths = paths
	}

	return v3, nil
}
