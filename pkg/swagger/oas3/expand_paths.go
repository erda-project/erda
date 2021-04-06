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
	"github.com/pkg/errors"
)

// ExpandPaths 展开所有路径中的所有请求体和响应体
func ExpandPaths(oas3 *openapi3.Swagger) error {
	if len(oas3.Paths) == 0 {
		return nil
	}
	for path_, pathItem := range oas3.Paths {
		for _, operation := range []*openapi3.Operation{
			pathItem.Connect, pathItem.Delete, pathItem.Get,
			pathItem.Head, pathItem.Options, pathItem.Patch,
			pathItem.Post, pathItem.Put, pathItem.Trace,
		} {
			if err := ExpandOperation(operation, oas3); err != nil {
				return errors.Wrapf(err, "failed to ExpandOperation, path: %s", path_)
			}
		}
	}
	return nil
}
