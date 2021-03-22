package oas3

import (
	"github.com/getkin/kin-openapi/openapi3"
)

func LoadFromData(data []byte) (*openapi3.Swagger, error) {
	return openapi3.NewSwaggerLoader().LoadSwaggerFromData(data)
}
