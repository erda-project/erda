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

package oas3

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ExpandOperation 展开所有 operation 中的请求体和响应体
func ExpandOperation(operation *openapi3.Operation, oas3 *openapi3.Swagger) error {
	if operation == nil {
		return nil
	}
	if len(operation.Parameters) > 0 {
		for _, parameterRef := range operation.Parameters {
			if parameterRef.Value == nil || parameterRef.Value.Schema == nil {
				continue
			}
			if err := ExpandSchemaRef(parameterRef.Value.Schema, oas3); err != nil {
				return err
			}
		}
	}
	if operation.RequestBody != nil {
		if err := ExpandRequestBody(operation.RequestBody, oas3); err != nil {
			return err
		}
	}
	if len(operation.Responses) != 0 {
		if err := ExpandResponses(operation.Responses, oas3); err != nil {
			return err
		}
	}
	return nil
}
