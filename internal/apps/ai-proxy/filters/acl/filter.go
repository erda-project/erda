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

package acl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-infra/base/logs"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "acl"
)

var (
	_ reverseproxy.RequestFilter = (*ACL)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type ACL struct {
	sources map[string]struct{}
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	var sources = make(map[string]struct{})
	for _, source := range cfg.Sources {
		sources[source] = struct{}{}
	}
	return &ACL{sources: sources}, nil
}

func (f *ACL) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)

	// only access control request rom platform erda.cloud
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyCredential{})
	if !ok || value == nil {
		return reverseproxy.Continue, nil
	}
	credential, ok := value.(*models.AIProxyCredentials)
	if !ok || credential == nil || credential.Platform != "erda.cloud" {
		return reverseproxy.Continue, nil
	}

	if source := infor.Header().Get(vars.XAIProxySource); source != "" {
		if _, ok := f.sources[source]; !ok {
			return reverseproxy.Continue, nil
		}
	}
	orgId := infor.Header().Get("Org-Id")
	if orgId == "" {
		l.Errorf("failed to get Org-Id from request header: Org-Id is missing or empty")
		http.Error(w, "Org-Id is missing or empty", http.StatusBadRequest)
		return reverseproxy.Intercept, nil
	}

	var orgServer = ctx.Value(vars.CtxKeyOrgSvc{}).(orgpb.OrgServiceServer)
	org, err := orgServer.GetOrg(apis.WithInternalClientContext(ctx, "ai-proxy"), &orgpb.GetOrgRequest{IdOrName: orgId})
	if err != nil {
		l.Errorf("failed to GetOrg, err: %v", err)
		http.Error(w, "failed to get org: "+err.Error(), http.StatusInternalServerError)
		return reverseproxy.Intercept, nil
	}
	if org.GetData() == nil {
		l.Errorf("failed to org.GetData: it is nil, err: %v", err)
		http.Error(w, "failed to get org", http.StatusBadRequest)
		return reverseproxy.Intercept, nil
	}
	config := org.GetData().GetConfig()
	if config == nil {
		l.Errorf("failed to org.GetData().GetConfig: it is nil, err: %v", err)
		http.Error(w, "failed to get org config", http.StatusInternalServerError)
		return reverseproxy.Intercept, errors.Wrap(err, "nil org config in gRPC response")
	}
	if !config.GetEnableAI() {
		l.Debugf("The org %s (%+v) doesn't enable AI Service, config: %+v", org.GetData().GetName(), org, config)
		w.Header().Set("Server", "AI Service on Erda")
		http.Error(w, fmt.Sprintf("The organization %s does not enable AI service", org.GetData().GetName()), http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}
	return reverseproxy.Continue, nil
}

func (f *ACL) Dependencies() map[string]json.RawMessage {
	return map[string]json.RawMessage{
		"context": {},
	}
}

type Config struct {
	Sources []string `json:"sources" yaml:"sources"`
}
