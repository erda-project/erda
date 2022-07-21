// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dto_test

import (
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
)

func TestUpstreamApiDto_Init(t *testing.T) {
	var apiDto dto.UpstreamApiDto
	if err := apiDto.Init(); err == nil {
		t.Fatal("apiDto should be invalid")
	} else {
		t.Log(err)
	}

	apiDto.Path = "some/uniform"
	if err := apiDto.Init(); err == nil {
		t.Fatal("apiDto should be invalid")
	} else {
		t.Log(err)
	}

	apiDto.Path = "/" + apiDto.Path
	if err := apiDto.Init(); err == nil {
		t.Fatal("apiDto should be invalid")
	} else {
		t.Log(err)
	}

	apiDto.Address = "some-service.svc.default.local"
	if err := apiDto.Init(); err == nil {
		t.Fatal("apiDto should be invalid")
	} else {
		t.Log(err)
	}

	apiDto.Address = "https://" + apiDto.Address
	if err := apiDto.Init(); err != nil {
		t.Fatalf("expects err==nil, got: %v", err)
	}
	if apiDto.Name != apiDto.Method+apiDto.Path {
		t.Fatal("apiDto.Name != apiDto.Method + apiDto.Path")
	}
	if apiDto.GatewayPath != apiDto.Path {
		t.Fatal("apiDto.GatewayPath != apiDto.Path")
	}

	gatewayPath := "/gateway/path"
	apiDto.GatewayPath = gatewayPath
	if err := apiDto.Init(); err != nil {
		t.Fatalf("expects err==nil, got: %v", err)
	}
	if apiDto.GatewayPath != gatewayPath {
		t.Fatalf("expects apiDto.GatewayPath: %s, got: %s", gatewayPath, apiDto.GatewayPath)
	}
}
