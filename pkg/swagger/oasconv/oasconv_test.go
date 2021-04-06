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

package oasconv_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/erda-project/erda/pkg/swagger/oas2"
	"github.com/erda-project/erda/pkg/swagger/oasconv"
)

const (
	oas2WithBasePath = "../testdata/oas2-with-basepath.json"
)

// go test -v -run TestOAS2ConvTo3
func TestOAS2ConvTo3(t *testing.T) {
	t.Run("testOas2ConvertTo3WithBasePath", testOas2ConvertTo3WithBasePath)
}

func testOas2ConvertTo3WithBasePath(t *testing.T) {
	data, err := ioutil.ReadFile(oas2WithBasePath)
	if err != nil {
		t.Fatalf("failed to ReadFile, filename: %s, err: %v", oas2WithBasePath, err)
	}

	v2, err := oas2.LoadFromData(data)
	if err != nil {
		t.Fatalf("failed to LoadFromData, filename: %s, err: %v", oas2WithBasePath, err)
	}

	v3, err := oasconv.OAS2ConvTo3(v2)
	if err != nil {
		t.Fatalf("failed to OAS2ConvTo3, filename: %s, err: %v", oas2WithBasePath, err)
	}

	if strings.HasPrefix(v2.BasePath, "/") {
		for k := range v2.Paths {
			if _, ok := v3.Paths[filepath.Join(v2.BasePath, k)]; !ok {
				t.Fatalf("failed to find the new path in openapi 3, v2.basePath: %s, current path: %s", v2.BasePath, k)
			}
		}
		for k := range v3.Paths {
			if !strings.HasPrefix(k, v2.BasePath) {
				t.Fatalf("the path in openapi3 do not has prefix v2.basePath, v2.basePath: %s, current path: %v", v2.BasePath, k)
			}
		}
	}
}
