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

package volcengine_viking

import (
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
)

func TestSignerOnProxyRequest_MetadataAKSK(t *testing.T) {
	sp := &serviceproviderpb.ServiceProvider{
		Type: common_types.ServiceProviderTypeVolcengineViking.String(),
		Metadata: &metadatapb.Metadata{
			Secret: map[string]*structpb.Value{
				"access_key_id": structpb.NewStringValue("ak-test"),
				"secret_key":    structpb.NewStringValue("sk-test"),
				"service":       structpb.NewStringValue("air"),
				"region":        structpb.NewStringValue("cn-north-1"),
				"tenant_id":     structpb.NewStringValue("tenant-a"),
			},
		},
	}

	req := httptest.NewRequest("POST", "https://api-knowledgebase.mlp.cn-beijing.volces.com/api/knowledge/service/rerank", strings.NewReader(`{"query":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	ctx := ctxhelper.InitCtxMapIfNeed(req.Context())
	ctxhelper.PutServiceProvider(ctx, sp)
	req = req.WithContext(ctx)

	out := req.Clone(ctx)
	pr := &httputil.ProxyRequest{In: req, Out: out}

	err := (&Signer{}).OnProxyRequest(pr)
	if err != nil {
		t.Fatalf("signer should not return error: %v", err)
	}
	if !strings.HasPrefix(pr.Out.Header.Get("Authorization"), "HMAC-SHA256 Credential=ak-test/") {
		t.Fatalf("unexpected Authorization: %s", pr.Out.Header.Get("Authorization"))
	}
	if pr.Out.Header.Get("X-Date") == "" {
		t.Fatalf("X-Date should be set")
	}
	if pr.Out.Header.Get("X-Content-Sha256") == "" {
		t.Fatalf("X-Content-Sha256 should be set")
	}
	if pr.Out.Header.Get("X-Tenant-Id") != "tenant-a" {
		t.Fatalf("tenant header should be set, got: %s", pr.Out.Header.Get("X-Tenant-Id"))
	}
}

func TestParseVikingSigningConfig_FromAPIKey(t *testing.T) {
	sp := &serviceproviderpb.ServiceProvider{
		ApiKey: strings.Join([]string{"ak-inline", "sk-inline"}, ":"),
	}
	cfg, ok := parseVikingSigningConfig(sp)
	if !ok {
		t.Fatalf("expected inline apiKey ak:sk to be parsed")
	}
	if cfg.ak != "ak-inline" || cfg.sk != "sk-inline" {
		t.Fatalf("unexpected parsed credentials: %#v", cfg)
	}
	if cfg.region != defaultSignRegion {
		t.Fatalf("unexpected default region: %s", cfg.region)
	}
	if cfg.service != defaultSignService {
		t.Fatalf("unexpected default service: %s", cfg.service)
	}
}
