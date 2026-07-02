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

package coordinator

import (
	"testing"

	"github.com/erda-project/erda/internal/apps/msp/resource/deploy/handlers"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func Test_isNeedDeployGatewayDependedAddon(t *testing.T) {
	type args struct {
		resourceSpecType string
		resourceSpecName string
		clusterConfig    map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_01",
			args: args{
				resourceSpecType: "addon",
				resourceSpecName: "api-gateway",
				clusterConfig: map[string]string{
					handlers.GatewayProviderVendorKey: "MSE",
				},
			},
			want: false,
		},
		{
			name: "Test_02",
			args: args{
				resourceSpecType: "addon",
				resourceSpecName: "api-gateway",
				clusterConfig: map[string]string{
					handlers.GatewayProviderVendorKey: "",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNeedDeployGatewayDependedAddon(tt.args.resourceSpecType, tt.args.resourceSpecName, tt.args.clusterConfig); got != tt.want {
				t.Errorf("isNeedDeployGatewayDependedAddon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasRealDeployment(t *testing.T) {
	info := &handlers.ResourceInfo{
		Dice: &diceyml.Object{
			AddOns: diceyml.AddOns{
				handlers.ResourceMysql: {
					Plan: "mysql:basic",
				},
			},
		},
	}

	if hasRealDeployment(info, true, true) {
		t.Fatalf("expected custom resource not to require real deployment")
	}
	if !hasRealDeployment(info, true, false) {
		t.Fatalf("expected resource with addons to require real deployment")
	}
	if hasRealDeployment(info, false, false) {
		t.Fatalf("expected resource without instance deployment not to require real deployment")
	}
}

func Test_shouldResolveDependencyResources(t *testing.T) {
	if shouldResolveDependencyResources(true, true, true) {
		t.Fatalf("expected custom resource dependencies not to be resolved")
	}
	if !shouldResolveDependencyResources(true, false, false) {
		t.Fatalf("expected tenant resources to resolve dependencies")
	}
	if !shouldResolveDependencyResources(false, true, false) {
		t.Fatalf("expected new instance resources to resolve dependencies")
	}
	if shouldResolveDependencyResources(false, false, false) {
		t.Fatalf("expected resources without tenant or instance work not to resolve dependencies")
	}
}
