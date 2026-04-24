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
	"encoding/json"
	"net/http/httputil"
	"strings"

	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	volcbase "github.com/volcengine/volc-sdk-golang/base"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

const (
	defaultSignRegion       = "cn-north-1"
	defaultSignService      = "air"
	defaultTenantHeaderName = "X-Tenant-Id"
)

func init() {
	filter_define.RegisterFilterCreator("volcengine-viking-rerank-signer", SignerCreator)
}

type Signer struct{}

var SignerCreator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Signer{}
}

func (f *Signer) Enable(pr *httputil.ProxyRequest) bool {
	return Enabled(pr.In.Context())
}

func (f *Signer) OnProxyRequest(pr *httputil.ProxyRequest) error {
	sp := ctxhelper.MustGetServiceProvider(pr.In.Context())
	signingCfg, ok := parseVikingSigningConfig(sp)
	// Backward compatibility: if only bearer api_key is configured, keep the old behavior.
	if !ok {
		return nil
	}

	if signingCfg.tenantID != "" {
		pr.Out.Header.Set(signingCfg.tenantHeader, signingCfg.tenantID)
	}

	// Ensure host is available for signer input.
	if pr.Out.Host == "" && pr.Out.URL != nil {
		pr.Out.Host = pr.Out.URL.Host
	}

	// Remove existing auth headers from director before re-signing.
	pr.Out.Header.Del("Authorization")
	pr.Out.Header.Del("X-Date")
	pr.Out.Header.Del("X-Content-Sha256")
	pr.Out.Header.Del("X-Security-Token")

	cred := volcbase.Credentials{
		AccessKeyID:     signingCfg.ak,
		SecretAccessKey: signingCfg.sk,
		Region:          signingCfg.region,
		Service:         signingCfg.service,
		SessionToken:    signingCfg.sessionToken,
	}
	cred.Sign(pr.Out)
	return nil
}

type vikingSigningConfig struct {
	ak           string
	sk           string
	region       string
	service      string
	sessionToken string
	tenantID     string
	tenantHeader string
}

func parseVikingSigningConfig(sp *serviceproviderpb.ServiceProvider) (vikingSigningConfig, bool) {
	m := metadata.FromProtobuf(sp.Metadata)

	get := func(keys ...string) string {
		for _, key := range keys {
			if v, ok := m.GetValueByKey(key, metadata.Config{IgnoreCase: true}); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}

	cfg := vikingSigningConfig{
		ak:           get("access_key_id", "access_key", "ak"),
		sk:           get("secret_access_key", "secret_key", "sk"),
		region:       get("region", "sign_region"),
		service:      get("service", "sign_service"),
		sessionToken: get("session_token", "security_token"),
		tenantID:     get("tenant_id", "tenant"),
		tenantHeader: get("tenant_header", "tenant_header_key"),
	}

	// fallback: parse apiKey as "ak:sk[:region[:service]]"
	if cfg.ak == "" || cfg.sk == "" {
		parts := strings.Split(sp.ApiKey, ":")
		if len(parts) >= 2 {
			if cfg.ak == "" {
				cfg.ak = strings.TrimSpace(parts[0])
			}
			if cfg.sk == "" {
				cfg.sk = strings.TrimSpace(parts[1])
			}
			if len(parts) >= 3 && cfg.region == "" {
				cfg.region = strings.TrimSpace(parts[2])
			}
			if len(parts) >= 4 && cfg.service == "" {
				cfg.service = strings.TrimSpace(parts[3])
			}
		}
	}

	if cfg.region == "" {
		cfg.region = defaultSignRegion
	}
	if cfg.service == "" {
		cfg.service = defaultSignService
	}
	if cfg.tenantHeader == "" {
		cfg.tenantHeader = defaultTenantHeaderName
	}

	return cfg, cfg.ak != "" && cfg.sk != ""
}
