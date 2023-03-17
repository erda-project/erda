// Copyright (c) 2023 Terminus, Inc.
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
