package oasconv

import (
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

	if v2.BasePath != "" {
		v3.Servers = append(v3.Servers, &openapi3.Server{
			URL:         v2.BasePath,
			Description: "",
			Variables:   nil,
		})
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

	return
}
