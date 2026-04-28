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
	"fmt"
	"net/http/httputil"
	"strings"

	volcbase "github.com/volcengine/volc-sdk-golang/base"

	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
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
	signingCfg, err := parseVikingSigningConfig(sp)
	if err != nil {
		return err
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
	// Volcengine signer signs all X-* headers by default.
	// Remove unstable proxy headers (for example X-Forwarded-For, X-Request-Id)
	// to avoid signature mismatch after gateway rewriting.
	sanitizeVolcSignHeaders(pr, signingCfg.tenantHeader)

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

func sanitizeVolcSignHeaders(pr *httputil.ProxyRequest, tenantHeader string) {
	if pr == nil || pr.Out == nil {
		return
	}
	tenantHeader = strings.TrimSpace(tenantHeader)
	for k := range pr.Out.Header {
		if !strings.HasPrefix(strings.ToLower(k), "x-") {
			continue
		}
		if strings.EqualFold(k, "X-Security-Token") {
			continue
		}
		if tenantHeader != "" && strings.EqualFold(k, tenantHeader) {
			continue
		}
		pr.Out.Header.Del(k)
	}
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

func parseVikingSigningConfig(sp *serviceproviderpb.ServiceProvider) (vikingSigningConfig, error) {
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
	if cfg.ak == "" || cfg.sk == "" {
		return vikingSigningConfig{}, fmt.Errorf("volcengine-viking requires metadata.secret.access_key_id and metadata.secret.secret_access_key")
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

	return cfg, nil
}
