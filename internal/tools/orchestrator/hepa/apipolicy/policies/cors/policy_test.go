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

package cors

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
	annotationscommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	mseCommon "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
)

func TestPolicy_setIngressAnnotations(t *testing.T) {
	type fields struct {
		BasePolicy apipolicy.BasePolicy
	}
	type args struct {
		gatewayProvider string
		policyDto       *PolicyDto
		locationSnippet string
	}

	corsEnable := "true"
	corsMethods := "GET, PUT, POST, DELETE, PATCH, OPTIONS"
	corsHeaders := "$http_access_control_request_headers"
	corsOrigin := "$from_request_origin_or_referer"
	corsCredentials := "true"
	corsMaxage := "86400"
	lSnippet := "xxxxx"

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *apipolicy.IngressAnnotation
	}{
		{
			name:   "Test_01",
			fields: fields{BasePolicy: apipolicy.BasePolicy{PolicyName: "cors"}},
			args: args{
				gatewayProvider: "MSE",
				policyDto: &PolicyDto{
					BaseDto:     apipolicy.BaseDto{},
					Methods:     "GET, PUT, POST, DELETE, PATCH, OPTIONS",
					Headers:     "$http_access_control_request_headers",
					Origin:      "$from_request_origin_or_referer",
					Credentials: true,
					MaxAge:      86400,
				},
				locationSnippet: "xxxxx",
			},
			want: &apipolicy.IngressAnnotation{
				Annotation: map[string]*string{
					string(annotationscommon.AnnotationEnableCORS):           &corsEnable,
					string(annotationscommon.AnnotationCORSAllowMethods):     &corsMethods,
					string(annotationscommon.AnnotationCORSAllowHeaders):     &corsHeaders,
					string(annotationscommon.AnnotationCORSAllowOrigin):      &corsOrigin,
					string(annotationscommon.AnnotationCORSAllowCredentials): &corsCredentials,
					string(annotationscommon.AnnotationCORSMaxAge):           &corsMaxage,
				},
			},
		},
		{
			name:   "Test_02",
			fields: fields{BasePolicy: apipolicy.BasePolicy{PolicyName: "cors"}},
			args: args{
				gatewayProvider: "",
				policyDto: &PolicyDto{
					BaseDto:     apipolicy.BaseDto{},
					Methods:     "GET, PUT, POST, DELETE, PATCH, OPTIONS",
					Headers:     "$http_access_control_request_headers",
					Origin:      "$from_request_origin_or_referer",
					Credentials: true,
					MaxAge:      86400,
				},
				locationSnippet: "xxxxx",
			},
			want: &apipolicy.IngressAnnotation{
				Annotation: map[string]*string{
					string(annotationscommon.AnnotationEnableCORS):           nil,
					string(annotationscommon.AnnotationCORSAllowMethods):     nil,
					string(annotationscommon.AnnotationCORSAllowHeaders):     nil,
					string(annotationscommon.AnnotationCORSAllowOrigin):      nil,
					string(annotationscommon.AnnotationCORSAllowCredentials): nil,
					string(annotationscommon.AnnotationCORSMaxAge):           nil,
				},
				LocationSnippet: &lSnippet,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := Policy{
				BasePolicy: tt.fields.BasePolicy,
			}
			if got := policy.setIngressAnnotations(tt.args.gatewayProvider, tt.args.policyDto, tt.args.locationSnippet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setIngressAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolicy_CreateDefaultConfig(t *testing.T) {
	type fields struct {
		BasePolicy apipolicy.BasePolicy
	}
	type args struct {
		gatewayProvider string
		ctx             map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   apipolicy.PolicyDto
	}{
		{
			name: "Test_01",
			fields: fields{
				BasePolicy: apipolicy.BasePolicy{
					PolicyName: apipolicy.Policy_Engine_CORS,
				},
			},
			args: args{
				gatewayProvider: mseCommon.MseProviderName,
				ctx:             nil,
			},
			want: &PolicyDto{
				Methods:     "GET, PUT, POST, DELETE, PATCH, OPTIONS",
				Headers:     "",
				Origin:      "",
				Credentials: true,
				MaxAge:      86400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := Policy{
				BasePolicy: tt.fields.BasePolicy,
			}
			if got := policy.CreateDefaultConfig(tt.args.gatewayProvider, tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateDefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
