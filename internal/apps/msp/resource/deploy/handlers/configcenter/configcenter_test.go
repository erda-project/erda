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

package configcenter

import (
	"testing"

	"github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestResetAddonsKeepsOriginalDependenciesWhenMseConfigured(t *testing.T) {
	p := &provider{DefaultDeployHandler: &handlers.DefaultDeployHandler{}}
	info := &handlers.ResourceInfo{
		Dice: &diceyml.Object{
			AddOns: diceyml.AddOns{
				handlers.ResourceNacos: {
					Plan: "nacos:basic",
				},
			},
		},
	}

	p.ResetAddons(info, map[string]string{
		handlers.MseNacosHost: "mse.example.com",
		handlers.MseNacosPort: "8848",
	})

	if _, ok := info.Dice.AddOns[handlers.ResourceNacos]; !ok {
		t.Fatalf("expected nacos dependency to be preserved")
	}
	if _, ok := info.Dice.AddOns[handlers.ResourceMSENacos]; ok {
		t.Fatalf("expected mse-nacos dependency not to be injected")
	}
}
