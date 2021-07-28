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

package apis_test

import (
	"testing"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/apis/dop"
	"github.com/erda-project/erda/pkg/swagger/oas3"
)

type StructA struct {
}

func TestStruct2OpenapiSchema(t *testing.T) {
	v3 := apis.NewSwagger("test")
	if err := dop.CreateAPIAssetVersion.AddOperationTo(v3); err != nil {
		t.Fatal(err)
	}
	data, err := oas3.MarshalYaml(v3)
	if err != nil {
		t.Fatalf("failed to MarshalYaml: %v", err)
	}
	t.Log(string(data))
}
