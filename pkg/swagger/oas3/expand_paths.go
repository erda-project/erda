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
