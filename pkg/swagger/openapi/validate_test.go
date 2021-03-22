package openapi_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"

	"github.com/erda-project/erda/pkg/swagger/openapi"
)

const testFile = "./testdata/portal.json"

// go test -v -run TestValidateOAS3
func TestValidateOAS3(t *testing.T) {
	data, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to ReadFile, err: %v", err)
	}

	swagger := new(openapi2.Swagger)
	if err = swagger.UnmarshalJSON(data); err != nil {
		t.Errorf("failed to UnmarshalJSON, err: %v", err)
	}

	oas3, err := openapi2conv.ToV3Swagger(swagger)
	if err != nil {
		t.Errorf("failed to ToV3Swagger, err: %v", err)
	}

	if err = openapi.ValidateOAS3(context.Background(), *oas3); err != nil {
		t.Errorf("failed to ValidateOAS3, err: %v", err)
	}
}
