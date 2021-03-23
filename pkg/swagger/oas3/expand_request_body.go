package oas3

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

// ExpandRequestBody 如果 request body 中引用了 Components, 则将 request body 的结构展开, 直到不再包含任何引用的结构
func ExpandRequestBody(ref *openapi3.RequestBodyRef, oas3 *openapi3.Swagger) error {
	if ref == nil {
		return errors.New("RequestBodyRef is nil")
	}
	if oas3 == nil {
		return errors.New("Swagger is nil")
	}

	if ref.Value == nil {
		return nil
	}

	if len(ref.Value.Content) == 0 {
		return nil
	}

	for mediaTypeName, mediaType := range ref.Value.Content {
		if mediaType.Schema == nil {
			continue
		}
		if err := ExpandSchemaRef(mediaType.Schema, oas3); err != nil {
			return errors.Wrapf(err, "failed to ExpandSchemaRef, mediaType: %s", mediaTypeName)
		}
	}

	return nil
}

func dedupeStringSlice(ss []string) []string {
	var (
		m      = make(map[string]bool, 0)
		result []string
	)
	for _, s := range ss {
		if _, ok := m[s]; !ok {
			result = append(result, s)
			m[s] = true
		}
	}
	return result
}
