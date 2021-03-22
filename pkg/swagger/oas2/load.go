package oas2

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi2"
)

func LoadFromData(data []byte) (*openapi2.Swagger, error) {
	var v2 openapi2.Swagger
	if err := json.Unmarshal(data, &v2); err != nil {
		return nil, err
	}
	return &v2, nil
}
