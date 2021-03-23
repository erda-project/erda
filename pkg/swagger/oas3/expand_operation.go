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
