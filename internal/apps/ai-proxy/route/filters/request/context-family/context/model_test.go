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

package context

import (
	"net/http"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func Test_buildLabels(t *testing.T) {
	labelsValue, _ := structpb.NewValue(map[string]any{"country": "JP", "template": "gpt-4o-chat"})
	countryValue, _ := structpb.NewValue("US")
	locationValue, _ := structpb.NewValue("earth")
	regionValue, _ := structpb.NewValue("us-east-1")
	providerCountryValue, _ := structpb.NewValue("JP")
	providerLocationValue, _ := structpb.NewValue("mars")
	providerRegionValue, _ := structpb.NewValue("ap-southeast-1")

	model := &modelpb.Model{
		Id:         "m1",
		Name:       "gpt-4o-azure",
		Publisher:  "openai",
		TemplateId: "tpl-1",
		Metadata: &metadatapb.Metadata{
			Public: map[string]*structpb.Value{
				"country":  countryValue,
				"location": locationValue,
				"region":   regionValue,
				"labels":   labelsValue,
			},
		},
	}
	provider := &providerpb.ServiceProvider{
		Type: "azure",
		Metadata: &metadatapb.Metadata{
			Public: map[string]*structpb.Value{
				"country":  providerCountryValue,
				"location": providerLocationValue,
				"region":   providerRegionValue,
			},
		},
	}

	result := policygroup.BuildLabels(model, provider)

	if result[common_types.PolicyLabelKeyModelInstanceID] != "m1" {
		t.Fatalf("expected instance-id m1, got %s", result[common_types.PolicyLabelKeyModelInstanceID])
	}
	if result[common_types.PolicyLabelKeyServiceProviderType] != "azure" {
		t.Fatalf("expected provider-type azure, got %s", result[common_types.PolicyLabelKeyServiceProviderType])
	}
	if result["country"] != "US" {
		t.Fatalf("expected country US from metadata, got %s", result["country"])
	}
	if result[common_types.PolicyLabelKeyTemplate] != "tpl-1" {
		t.Fatalf("expected template tpl-1, got %s", result[common_types.PolicyLabelKeyTemplate])
	}
	if result[common_types.PolicyLabelKeyTemplate] == "" || result[common_types.PolicyLabelKeyModelPublisherModelTemplateID] == "" {
		t.Fatalf("expected publisher-model and template labels populated")
	}
}

func Test_buildRequestMeta(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Request-Id", "req-1")
	headers.Set("X-AI-Proxy-Generated-Call-Id", "call-1")
	headers.Set("User-Id", "u1")

	meta := policygroup.BuildRequestMetaFromHeader(headers)
	if meta.Keys[common_types.StickyKeyOfXRequestID] != "req-1" {
		t.Fatalf("expected x-request-id to be lowercased, got %s", meta.Keys[common_types.StickyKeyOfXRequestID])
	}
	if meta.Keys[common_types.StickyKeyPrefixFromReqHeader+"user-id"] != "u1" {
		t.Fatalf("expected user-id to be kept, got %s", meta.Keys[common_types.StickyKeyPrefixFromReqHeader+"user-id"])
	}
	if meta.Keys[common_types.StickyKeyPrefixFromReqHeader+"x-ai-proxy-generated-call-id"] != "call-1" {
		t.Fatalf("expected call id captured, got %s", meta.Keys[common_types.StickyKeyPrefixFromReqHeader+"x-ai-proxy-generated-call-id"])
	}
}
