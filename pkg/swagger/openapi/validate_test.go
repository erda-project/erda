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

package openapi_test

//import (
//	"context"
//	"io/ioutil"
//	"testing"
//
//	"github.com/getkin/kin-openapi/openapi2"
//	"github.com/getkin/kin-openapi/openapi2conv"
//
//	"github.com/erda-project/erda/pkg/swagger/openapi"
//)
//
//const testFile = "./testdata/portal.json"
//
//// go test -v -run TestValidateOAS3
//func TestValidateOAS3(t *testing.T) {
//	data, err := ioutil.ReadFile(testFile)
//	if err != nil {
//		t.Errorf("failed to ReadFile, err: %v", err)
//	}
//
//	swagger := new(openapi2.Swagger)
//	if err = swagger.UnmarshalJSON(data); err != nil {
//		t.Errorf("failed to UnmarshalJSON, err: %v", err)
//	}
//
//	oas3, err := openapi2conv.ToV3Swagger(swagger)
//	if err != nil {
//		t.Errorf("failed to ToV3Swagger, err: %v", err)
//	}
//
//	if err = openapi.ValidateOAS3(context.Background(), *oas3); err != nil {
//		t.Errorf("failed to ValidateOAS3, err: %v", err)
//	}
//}
