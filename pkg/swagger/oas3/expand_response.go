// 注意 本 package 下所有 Expand* 函数, 都会修改入参 oas3 *openapi.Swagger 的内部结构,
// 如果 oas3 中的被引类型不是展开的, 也会被一并展开

package oas3

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

// ExpandResponses 将所有的 response body 都展开
func ExpandResponses(responses openapi3.Responses, oas3 *openapi3.Swagger) error {
	for statusCode, responseRef := range responses {
		if responseRef.Value == nil {
			continue
		}

		for mediaTypeName, mediaType := range responseRef.Value.Content {
			if mediaType.Schema == nil {
				continue
			}
			if err := ExpandSchemaRef(mediaType.Schema, oas3); err != nil {
				return errors.Wrapf(err, "failed to ExpandSchemaRef, statusCode: %s, mediaType: %s", statusCode, mediaTypeName)
			}
		}
	}

	return nil
}
