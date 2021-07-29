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
	"context"
	"testing"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/apis/dop"
	"github.com/erda-project/erda/pkg/swagger/oas3"
)

func TestApiSpec_AddOperationTo(t *testing.T) {
	var testAPIs = []*apis.ApiSpec{
		&dop.CreateAPIAssetVersion,
		&dop.GetAccess,
	}
	v3 := apis.NewSwagger("test")
	for _, api := range testAPIs {
		if err := api.AddOperationTo(v3); err != nil {
			t.Fatal(err)
		}
	}
	data, err := oas3.MarshalYaml(v3)
	if err != nil {
		t.Fatal(err)
	}
	newV3, err := oas3.LoadFromData(data)
	if err != nil {
		t.Fatal(err)
	}
	if err = oas3.ValidateOAS3(context.Background(), *newV3); err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
