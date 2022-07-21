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
	"strings"
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

func TestUpstreamRegisterDto_Init(t *testing.T) {
	jsonstr := `{"registerId":123456}`
	type fields struct {
		OldRegisterId int `json:"registerId"`
	}
	obj := &fields{}
	_ = json.Unmarshal([]byte(jsonstr), obj)
	t.Logf("%+v", obj)
}

func TestUpstreamApiDto_Init2(t *testing.T) {
	var d = getCase()
	if err := d.Init(); err != nil {
		t.Fatalf("expects err==nil, got: %v", err)
	}

	t.Run("default scene", func(t *testing.T) {
		var d = getCase()
		d.Scene = ""
		_ = d.Init()
		if d.Scene != orm.UnityScene {
			t.Fatalf("expects d.Scene: %s, got: %s", orm.UnityScene, d.Scene)
		}
	})
	t.Run("invalid scene", func(t *testing.T) {
		var d = getCase()
		d.Scene = "some-scene"
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})
	t.Run("invalid appName", func(t *testing.T) {
		var d = getCase()
		d.AppName = ""
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})
	t.Run("invalid orgId", func(t *testing.T) {
		var d = getCase()
		d.OrgId = ""
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})
	t.Run("invalid projectId", func(t *testing.T) {
		var d = getCase()
		d.ProjectId = ""
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})
	t.Run("invalid apiList", func(t *testing.T) {
		var d = getCase()
		d.ApiList = nil
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})
	t.Run("invalid env - 0", func(t *testing.T) {
		var d = getCase()
		d.Env = ""
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})
	t.Run("invalid env - 1", func(t *testing.T) {
		var d = getCase()
		d.Env = "some-env"
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})
	t.Run("invalid registerId", func(t *testing.T) {
		var d = getCase()
		d.RegisterId = ""
		if err := d.Init(); err == nil {
			t.Fatalf("expects err != nil, got nil")
		}
	})

	t.Run("default DiceService", func(t *testing.T) {
		var d = getCase()
		var serviceName = "meta-store"
		d.DiceService = ""
		d.ServiceAlias = serviceName
		_ = d.Init()
		if d.DiceService != serviceName {
			t.Fatalf("expects d.DcieService: %s, got: %s", serviceName, d.DiceService)
		}
	})
	t.Run("default UpstreamName - 0", func(t *testing.T) {
		var d = getCase()
		d.DiceService = ""
		d.RuntimeId = ""
		_ = d.Init()
		if expects := strings.Join([]string{d.AppName, d.ServiceAlias}, "/"); d.UpstreamName != expects {
			t.Fatalf("expects UpstreamName: %s, got: %s", expects, d.UpstreamName)
		}
	})
	t.Run("default UpstreamName - 1", func(t *testing.T) {
		var d = getCase()
		d.DiceService = ""
		_ = d.Init()
		if expects := strings.Join([]string{d.AppName, d.ServiceAlias, d.RuntimeId}, "/"); d.UpstreamName != expects {
			t.Fatalf("expects UpstreamName: %s, got: %s", expects, d.UpstreamName)
		}
	})
	t.Run("default Upstream - 2: hub scene", func(t *testing.T) {
		var d = getCase()
		d.DiceService = ""
		d.Scene = orm.HubScene
		_ = d.Init()
		if expects := strings.Join([]string{
			"custom-registered",
			d.AppName,
			d.ServiceAlias,
			d.RuntimeId,
		}, "/"); d.UpstreamName != expects {
			t.Fatalf("expects UpstreamName: %s, got: %s", expects, d.UpstreamName)
		}
	})
	t.Run("default PathPrefix - 0", func(t *testing.T) {
		var d = getCase()
		_ = d.Init()
		if d.PathPrefix != nil {
			t.Fatalf("expects PathPrefix: nil, got: %v", d.PathPrefix)
		}
	})
	t.Run("default PathPrefix - 1", func(t *testing.T) {
		var d = getCase()
		d.RuntimeName = ""
		_ = d.Init()
		if expects := "/" + strings.Join([]string{d.AppName, d.ServiceAlias, d.RuntimeId}, "/"); d.PathPrefix == nil {
			t.Fatalf("expects PathPrefix: %s, got: %v", expects, d.PathPrefix)
		} else if *d.PathPrefix != expects {
			t.Fatalf("expects PathPrefix: %s, got: %v", expects, *d.PathPrefix)
		}
	})
}

func TestUpstreamRegisterDto_AdjustAPIsDomains(t *testing.T) {
	t.Run("domains in unity scene", func(t *testing.T) {
		var d = getCase()
		d.Scene = orm.UnityScene
		if err := d.AdjustAPIsDomains(); err != nil {
			t.Fatalf("expects err==nil, got: %v", err)
		}
		for i := range d.ApiList {
			if domain := d.ApiList[i].Domain; domain != "" {
				t.Fatalf("expects d.ApiList[%v].Domain==\"\", got: %s", i, domain)
			}
		}
	})

	t.Run("domains in other scene", func(t *testing.T) {
		var d = getCase()
		d = getCase()
		d.Scene = orm.HubScene
		d.ApiList = append(d.ApiList, dto.UpstreamApiDto{
			Path:        "/api/load-data/v2",
			GatewayPath: "",
			Method:      "",
			Address:     "",
			IsInner:     false,
			IsCustom:    true,
			Doc:         nil,
			Name:        "",
			Domain:      "erda.cloud",
		})
		if err := d.AdjustAPIsDomains(); err == nil {
			t.Log(d.ApiList)
			t.Fatal("expects err!=nil, got: nil")
		} else {
			t.Log(err)
		}
	})

}

func getCase() dto.UpstreamRegisterDto {
	return dto.UpstreamRegisterDto{
		Az:           "erda-0",
		UpstreamName: "meta-store",
		DiceAppId:    "34",
		DiceService:  "meta-store-service",
		RuntimeName:  "meta-store-runtime",
		RuntimeId:    "56",
		AppName:      "trans-app",
		ServiceAlias: "meta-store-service",
		OrgId:        "1",
		ProjectId:    "2",
		Env:          "TEST",
		ApiList: []dto.UpstreamApiDto{
			{
				Path:        "/api/load-data",
				GatewayPath: "",
				Method:      "",
				Address:     "",
				IsInner:     false,
				IsCustom:    true,
				Doc:         nil,
				Name:        "",
				Domain:      "baidu.com,google.com",
			},
		},
		OldRegisterId: nil,
		RegisterId:    "xxx-yyy-zzz",
		PathPrefix:    nil,
		Scene:         orm.UnityScene,
	}
}
