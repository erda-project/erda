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
	"encoding/json"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/kong/dto"
	"testing"
)

func TestKongTargetDto_JSONUnmarshal(t *testing.T) {
	var body = `{"created_at":1679018347.926,"id":"39c8ae64-5b4c-476c-a44d-1c850c00d433","tags":null,"weight":100,"target":"172.16.142.209:8080","upstream":{"id":"397a3af9-f77b-4e1c-92f3-2dfef6034229"}}`
	var data dto.KongTargetDto
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", data)
	t.Logf("createdAt: %v", data.GetCreatedAt())
}
