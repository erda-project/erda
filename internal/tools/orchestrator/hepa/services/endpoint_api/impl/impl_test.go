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

package impl

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/api_policy"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/domain"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/endpoint_api"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/global"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/micro_api"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/openapi_consumer"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/openapi_rule"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/zone"
)

func TesthubExists(t *testing.T) {

}

func TestGatewayOpenapiServiceImpl_touchServiceForExternalService(t *testing.T) {
	type fields struct {
		packageDb       service.GatewayPackageService
		packageApiDb    service.GatewayPackageApiService
		zoneInPackageDb service.GatewayZoneInPackageService
		apiInPackageDb  service.GatewayApiInPackageService
		packageInDb     service.GatewayPackageInConsumerService
		serviceDb       service.GatewayServiceService
		routeDb         service.GatewayRouteService
		consumerDb      service.GatewayConsumerService
		apiDb           service.GatewayApiService
		upstreamApiDb   service.GatewayUpstreamApiService
		azDb            service.GatewayAzInfoService
		kongDb          service.GatewayKongInfoService
		hubInfoDb       service.GatewayHubInfoService
		apiBiz          *micro_api.GatewayApiService
		zoneBiz         *zone.GatewayZoneService
		ruleBiz         *openapi_rule.GatewayOpenapiRuleService
		consumerBiz     *openapi_consumer.GatewayOpenapiConsumerService
		globalBiz       *global.GatewayGlobalService
		policyBiz       *api_policy.GatewayApiPolicyService
		runtimeDb       service.GatewayRuntimeServiceService
		domainBiz       *domain.GatewayDomainService
		ctx             context.Context
		reqCtx          context.Context
	}
	type args struct {
		info endpoint_api.PackageApiInfo
		z    orm.GatewayZone
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *corev1.Service
		wantErr bool
	}{
		{
			name:   "Test_01",
			fields: fields{},
			args: args{
				info: endpoint_api.PackageApiInfo{
					GatewayPackageApi: &orm.GatewayPackageApi{
						PackageId:        "066523a826ac4e81afa908a1f1e25115",
						ApiPath:          "/test/for-inner/urls",
						Method:           "",
						RedirectAddr:     "http://bbb-151f3a62c7.project-5846-test.svc.cluster.local:80/",
						RedirectPath:     "/",
						Description:      "",
						DiceApp:          "bbb",
						DiceService:      "bbb",
						AclType:          "",
						Origin:           "custom",
						DiceApiId:        "",
						RedirectType:     "url",
						RuntimeServiceId: "",
						ZoneId:           "1f242ca6d45e43d2b52124eec2138f4d",
						CloudapiApiId:    "",
						BaseRow: orm.BaseRow{
							Id:        "5ba4b3809d6143c9ac426f96756c1f04",
							IsDeleted: "N",
						},
					},
					Hosts:               []string{"bbb-151f3a62c7.project-5846-test.svc.cluster.local"},
					ProjectId:           "5846",
					Env:                 "test",
					Az:                  "test",
					InjectRuntimeDomain: false,
				},
				z: orm.GatewayZone{
					Name: "dice-test-5846-api-5ba4b3809d6143c9ac426f96756c1f04-1f2f4d",
					BaseRow: orm.BaseRow{
						Id:        "1f242ca6d45e43d2b52124eec2138f4d",
						IsDeleted: "N",
					},
				},
			},
			want: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dice-test-5846-api-5ba4b3809d6143c9ac426f96756c1f04-1f2f4d",
					Namespace: "project-5846-test",
					Labels: map[string]string{
						"packageId":                "066523a826ac4e81afa908a1f1e25115",
						"erda.gateway.projectId":   "5846",
						"erda.gateway.appName":     "bbb",
						"erda.gateway.serviceName": "bbb",
						"erda.gateway.workspace":   "test",
					},
				},
				Spec: corev1.ServiceSpec{
					ExternalName: "bbb-151f3a62c7.project-5846-test.svc.cluster.local",
					Type:         corev1.ServiceTypeExternalName,
					Ports: []corev1.ServicePort{
						{
							Name:       "target",
							Protocol:   corev1.ProtocolTCP,
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
					},
				},
				Status: corev1.ServiceStatus{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := GatewayOpenapiServiceImpl{
				packageDb:       tt.fields.packageDb,
				packageApiDb:    tt.fields.packageApiDb,
				zoneInPackageDb: tt.fields.zoneInPackageDb,
				apiInPackageDb:  tt.fields.apiInPackageDb,
				packageInDb:     tt.fields.packageInDb,
				serviceDb:       tt.fields.serviceDb,
				routeDb:         tt.fields.routeDb,
				consumerDb:      tt.fields.consumerDb,
				apiDb:           tt.fields.apiDb,
				upstreamApiDb:   tt.fields.upstreamApiDb,
				azDb:            tt.fields.azDb,
				kongDb:          tt.fields.kongDb,
				hubInfoDb:       tt.fields.hubInfoDb,
				apiBiz:          tt.fields.apiBiz,
				zoneBiz:         tt.fields.zoneBiz,
				ruleBiz:         tt.fields.ruleBiz,
				consumerBiz:     tt.fields.consumerBiz,
				globalBiz:       tt.fields.globalBiz,
				policyBiz:       tt.fields.policyBiz,
				runtimeDb:       tt.fields.runtimeDb,
				domainBiz:       tt.fields.domainBiz,
				ctx:             tt.fields.ctx,
				reqCtx:          tt.fields.reqCtx,
			}
			got, err := impl.touchServiceForExternalService(tt.args.info, tt.args.z)
			if (err != nil) != tt.wantErr {
				t.Errorf("touchServiceForExternalService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("touchServiceForExternalService() got = %v, want %v", got, tt.want)
			}
		})
	}
}
