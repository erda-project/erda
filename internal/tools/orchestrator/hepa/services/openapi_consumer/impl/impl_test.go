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
	"time"

	"bou.ke/monkey"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	orgCache "github.com/erda-project/erda/internal/tools/orchestrator/hepa/cache/org"
	gateway_providers "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse"
	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/openapi_rule"
)

func TestGatewayOpenapiConsumerServiceImpl_createMseGatewayCredential(t *testing.T) {
	type fields struct {
		packageDb      service.GatewayPackageService
		packageApiDb   service.GatewayPackageApiService
		consumerDb     service.GatewayConsumerService
		azDb           service.GatewayAzInfoService
		kongDb         service.GatewayKongInfoService
		packageInDb    service.GatewayPackageInConsumerService
		packageApiInDb service.GatewayPackageApiInConsumerService
		credentialDb   service.GatewayCredentialService
		ruleBiz        *openapi_rule.GatewayOpenapiRuleService
		reqCtx         context.Context
	}
	type args struct {
		req    *providerDto.CredentialReqDto
		config *providerDto.CredentialDto
	}

	consumerDb, _ := service.NewGatewayConsumerServiceImpl()
	credentialDb, _ := service.NewGatewayCredentialServiceImpl()
	redirecturl := []string{"/test01", "test02"}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				consumerDb:   consumerDb,
				credentialDb: credentialDb,
			},
			args: args{
				req: &providerDto.CredentialReqDto{
					ConsumerId: "b13878b8-9686-4a1c-a897-3bd5e34785ef",
					PluginName: "key-auth",
					Config: &providerDto.CredentialDto{
						ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						RedirectUrl:  redirecturl,
						RedirectUrls: []string{"/a", "/b"},
					},
				},
				config: &providerDto.CredentialDto{
					ConsumerId: "b13878b8-9686-4a1c-a897-3bd5e34785ef",
				},
			},
			wantErr: false,
		},
	}
	tTime := time.Now()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := GatewayOpenapiConsumerServiceImpl{
				packageDb:      tt.fields.packageDb,
				packageApiDb:   tt.fields.packageApiDb,
				consumerDb:     tt.fields.consumerDb,
				azDb:           tt.fields.azDb,
				kongDb:         tt.fields.kongDb,
				packageInDb:    tt.fields.packageInDb,
				packageApiInDb: tt.fields.packageApiInDb,
				credentialDb:   tt.fields.credentialDb,
				ruleBiz:        tt.fields.ruleBiz,
				reqCtx:         tt.fields.reqCtx,
			}
			monkey.Patch(orgCache.GetOrgByOrgID, func(orgID string) (*orgpb.Org, bool) {
				return &orgpb.Org{
					Name: "test",
				}, true
			})
			monkey.PatchInstanceMethod(reflect.TypeOf(impl.consumerDb), "GetByConsumerId", func(*service.GatewayConsumerServiceImpl, string) (*orm.GatewayConsumer, error) {
				return &orm.GatewayConsumer{
					ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
					ConsumerName: "abc",
					OrgId:        "633",
					ProjectId:    "5846",
					Env:          "TEST",
					Az:           "test",
					Description:  "ssss",
					BaseRow: orm.BaseRow{
						Id:         "22255a42f7a848619f9ffe0fa1fdf85b",
						IsDeleted:  "N",
						CreateTime: tTime,
						UpdateTime: tTime,
					},
				}, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(impl.credentialDb), "Insert", func(*service.GatewayCredentialServiceImpl, *orm.GatewayCredential) error {
				return nil
			})
			defer monkey.UnpatchAll()
			if err := impl.createMseGatewayCredential(tt.args.req, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("createMseGatewayCredential() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGatewayOpenapiConsumerServiceImpl_getCredentialList(t *testing.T) {
	type fields struct {
		packageDb      service.GatewayPackageService
		packageApiDb   service.GatewayPackageApiService
		consumerDb     service.GatewayConsumerService
		azDb           service.GatewayAzInfoService
		kongDb         service.GatewayKongInfoService
		packageInDb    service.GatewayPackageInConsumerService
		packageApiInDb service.GatewayPackageApiInConsumerService
		credentialDb   service.GatewayCredentialService
		ruleBiz        *openapi_rule.GatewayOpenapiRuleService
		reqCtx         context.Context
	}
	type args struct {
		gatewayAdapter gateway_providers.GatewayAdapter
		consumerId     string
	}

	credentialDb, _ := service.NewGatewayCredentialServiceImpl()
	tTime := time.Now()
	test01Result := make(map[string]providerDto.CredentialListDto)
	test01Result[orm.KEYAUTH] = providerDto.CredentialListDto{
		Total: 1,
		Data: []providerDto.CredentialDto{
			{
				ConsumerId: "b13878b8-9686-4a1c-a897-3bd5e34785ef",
				CreatedAt:  tTime.Unix() * 1000,
				Id:         "d2e23e27d5ac4f7e8f8d4bf83c6daf13",
				Key:        "dae6ece8afc24c9581172dfd95b298e4",
			},
		},
	}

	test01Result[orm.OAUTH2] = providerDto.CredentialListDto{
		Total: 1,
		Data: []providerDto.CredentialDto{
			{
				ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
				CreatedAt:    tTime.Unix() * 1000,
				Id:           "733d1f97b3f645e18b78bb7b7fb9792b",
				Name:         "App",
				RedirectUrl:  []string{"http://none"},
				ClientId:     "dae6ece8afc24c9581172dfd95b298e4",
				ClientSecret: "335698f06d2b4977b1060b034c3006a1",
			},
		},
	}

	test01Result[orm.SIGNAUTH] = providerDto.CredentialListDto{
		Total: 1,
		Data: []providerDto.CredentialDto{
			{
				ConsumerId: "b13878b8-9686-4a1c-a897-3bd5e34785ef",
				CreatedAt:  tTime.Unix() * 1000,
				Id:         "57ccd4e3d2464e9aa117ab69e7cf30de",
				Key:        "dae6ece8afc24c9581172dfd95b298e4",
				Secret:     "335698f06d2b4977b1060b034c3006a1",
			},
		},
	}

	test01Result[orm.HMACAUTH] = providerDto.CredentialListDto{
		Total: 1,
		Data: []providerDto.CredentialDto{
			{
				ConsumerId: "b13878b8-9686-4a1c-a897-3bd5e34785ef",
				CreatedAt:  tTime.Unix() * 1000,
				Id:         "a076fbd5c81d4a39bd371c4d6f396abd",
				Key:        "dae6ece8afc24c9581172dfd95b298e4",
				Secret:     "335698f06d2b4977b1060b034c3006a1",
				Username:   "",
			},
		},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]providerDto.CredentialListDto
		wantErr bool
	}{
		{
			name: "Test01",
			fields: fields{
				credentialDb: credentialDb,
			},
			args: args{
				gatewayAdapter: &mse.MseAdapterImpl{},
				consumerId:     "b13878b8-9686-4a1c-a897-3bd5e34785ef",
			},
			want:    test01Result,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := GatewayOpenapiConsumerServiceImpl{
				packageDb:      tt.fields.packageDb,
				packageApiDb:   tt.fields.packageApiDb,
				consumerDb:     tt.fields.consumerDb,
				azDb:           tt.fields.azDb,
				kongDb:         tt.fields.kongDb,
				packageInDb:    tt.fields.packageInDb,
				packageApiInDb: tt.fields.packageApiInDb,
				credentialDb:   tt.fields.credentialDb,
				ruleBiz:        tt.fields.ruleBiz,
				reqCtx:         tt.fields.reqCtx,
			}

			monkey.PatchInstanceMethod(reflect.TypeOf(impl.credentialDb), "SelectByConsumerId", func(*service.GatewayCredentialServiceImpl, string) ([]orm.GatewayCredential, error) {
				return []orm.GatewayCredential{
					{
						ConsumerId:       "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						ConsumerName:     "abc",
						PluginName:       "sign-auth",
						OrgId:            "633",
						ProjectId:        "5846",
						Env:              "TEST",
						Az:               "test",
						Key:              "dae6ece8afc24c9581172dfd95b298e4",
						Secret:           "335698f06d2b4977b1060b034c3006a1",
						KeepToken:        "Y",
						ClockSkewSeconds: "60",
						BaseRow: orm.BaseRow{
							Id:         "57ccd4e3d2464e9aa117ab69e7cf30de",
							IsDeleted:  "N",
							CreateTime: tTime,
						},
					},
					{
						ConsumerId:       "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						ConsumerName:     "abc",
						PluginName:       "oauth2",
						OrgId:            "633",
						ProjectId:        "5846",
						Env:              "TEST",
						Az:               "test",
						Key:              "dae6ece8afc24c9581172dfd95b298e4",
						Secret:           "335698f06d2b4977b1060b034c3006a1",
						KeepToken:        "Y",
						ClockSkewSeconds: "60",
						RedirectUrl:      "http://none",
						RedirectUrls:     "",
						Name:             "App",
						ClientId:         "dae6ece8afc24c9581172dfd95b298e4",
						ClientSecret:     "335698f06d2b4977b1060b034c3006a1",
						BaseRow: orm.BaseRow{
							Id:         "733d1f97b3f645e18b78bb7b7fb9792b",
							IsDeleted:  "N",
							CreateTime: tTime,
						},
					},
					{
						ConsumerId:       "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						ConsumerName:     "abc",
						PluginName:       "hmac-auth",
						OrgId:            "633",
						ProjectId:        "5846",
						Env:              "TEST",
						Az:               "test",
						Key:              "dae6ece8afc24c9581172dfd95b298e4",
						Secret:           "335698f06d2b4977b1060b034c3006a1",
						KeepToken:        "Y",
						ClockSkewSeconds: "60",
						BaseRow: orm.BaseRow{
							Id:         "a076fbd5c81d4a39bd371c4d6f396abd",
							IsDeleted:  "N",
							CreateTime: tTime,
						},
					},
					{
						ConsumerId:       "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						ConsumerName:     "abc",
						PluginName:       "key-auth",
						OrgId:            "633",
						ProjectId:        "5846",
						Env:              "TEST",
						Az:               "test",
						Key:              "dae6ece8afc24c9581172dfd95b298e4",
						KeepToken:        "Y",
						ClockSkewSeconds: "60",
						BaseRow: orm.BaseRow{
							Id:         "d2e23e27d5ac4f7e8f8d4bf83c6daf13",
							IsDeleted:  "N",
							CreateTime: tTime,
						},
					},
				}, nil
			})
			defer monkey.UnpatchAll()

			got, err := impl.getCredentialList(tt.args.gatewayAdapter, tt.args.consumerId)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCredentialList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCredentialList() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGatewayOpenapiConsumerServiceImpl_mseConsumerConfig(t *testing.T) {
	type fields struct {
		packageDb      service.GatewayPackageService
		packageApiDb   service.GatewayPackageApiService
		consumerDb     service.GatewayConsumerService
		azDb           service.GatewayAzInfoService
		kongDb         service.GatewayKongInfoService
		packageInDb    service.GatewayPackageInConsumerService
		packageApiInDb service.GatewayPackageApiInConsumerService
		credentialDb   service.GatewayCredentialService
		ruleBiz        *openapi_rule.GatewayOpenapiRuleService
		reqCtx         context.Context
	}
	type args struct {
		consumers []orm.GatewayConsumer
	}

	credentialDb, _ := service.NewGatewayCredentialServiceImpl()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []mseDto.Consumers
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				credentialDb: credentialDb,
			},
			args: args{
				consumers: []orm.GatewayConsumer{{
					ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
					ConsumerName: "abc",
					OrgId:        "633",
					ProjectId:    "5846",
					Env:          "TEST",
					Az:           "test",
					BaseRow: orm.BaseRow{
						Id: "22255a42f7a848619f9ffe0fa1fdf85b",
					},
				}},
			},
			want: []mseDto.Consumers{
				{
					Name:       "633.5846.TEST.test:abc",
					Credential: "dae6ece8afc24c9581172dfd95b298e4",
				},
			},
			wantErr: false,
		},
		{
			name: "Test_02",
			fields: fields{
				credentialDb: credentialDb,
			},
			args: args{
				consumers: []orm.GatewayConsumer{{
					ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
					ConsumerName: "abc",
					OrgId:        "633",
					ProjectId:    "5846",
					Env:          "TEST",
					Az:           "test",
					BaseRow: orm.BaseRow{
						Id: "22255a42f7a848619f9ffe0fa1fdf85b",
					},
				}},
			},
			want: []mseDto.Consumers{
				{
					Name:   "633.5846.TEST.test:abc",
					Key:    "dae6ece8afc24c9581172dfd95b298e4",
					Secret: "335698f06d2b4977b1060b034c3006a1",
				},
			},
			wantErr: false,
		},
		{
			name: "Test_03",
			fields: fields{
				credentialDb: credentialDb,
			},
			args: args{
				consumers: []orm.GatewayConsumer{{
					ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
					ConsumerName: "abc",
					OrgId:        "633",
					ProjectId:    "5846",
					Env:          "TEST",
					Az:           "test",
					BaseRow: orm.BaseRow{
						Id: "22255a42f7a848619f9ffe0fa1fdf85b",
					},
				}},
			},
			want: []mseDto.Consumers{
				{
					Name:             "633.5846.TEST.test:abc",
					FromParams:       []string{"test"},
					FromCookies:      []string{"test"},
					KeepToken:        true,
					ClockSkewSeconds: 30,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := GatewayOpenapiConsumerServiceImpl{
				packageDb:      tt.fields.packageDb,
				packageApiDb:   tt.fields.packageApiDb,
				consumerDb:     tt.fields.consumerDb,
				azDb:           tt.fields.azDb,
				kongDb:         tt.fields.kongDb,
				packageInDb:    tt.fields.packageInDb,
				packageApiInDb: tt.fields.packageApiInDb,
				credentialDb:   tt.fields.credentialDb,
				ruleBiz:        tt.fields.ruleBiz,
				reqCtx:         tt.fields.reqCtx,
			}
			monkey.PatchInstanceMethod(reflect.TypeOf(impl.credentialDb), "SelectByConsumerId", func(*service.GatewayCredentialServiceImpl, string) ([]orm.GatewayCredential, error) {
				switch tt.name {
				case "Test_02":
					return []orm.GatewayCredential{{
						ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						ConsumerName: "abc",
						PluginName:   "hmac-auth",
						OrgId:        "633",
						ProjectId:    "5846",
						Env:          "TEST",
						Az:           "test",
						Key:          "dae6ece8afc24c9581172dfd95b298e4",
						Secret:       "335698f06d2b4977b1060b034c3006a1",
					}}, nil

				case "Test_03":
					return []orm.GatewayCredential{{
						ConsumerId:       "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						ConsumerName:     "abc",
						PluginName:       "jwt-auth",
						OrgId:            "633",
						ProjectId:        "5846",
						Env:              "TEST",
						Az:               "test",
						Key:              "dae6ece8afc24c9581172dfd95b298e4",
						Secret:           "335698f06d2b4977b1060b034c3006a1",
						FromCookies:      "test",
						FromParams:       "test",
						KeepToken:        "Y",
						ClockSkewSeconds: "30",
					}}, nil
				default:
					return []orm.GatewayCredential{{
						ConsumerId:   "b13878b8-9686-4a1c-a897-3bd5e34785ef",
						ConsumerName: "abc",
						PluginName:   "key-auth",
						OrgId:        "633",
						ProjectId:    "5846",
						Env:          "TEST",
						Az:           "test",
						Key:          "dae6ece8afc24c9581172dfd95b298e4",
					}}, nil

				}
			})
			defer monkey.UnpatchAll()

			got, err := impl.mseConsumerConfig(tt.args.consumers)
			if (err != nil) != tt.wantErr {
				t.Errorf("mseConsumerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mseConsumerConfig() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
