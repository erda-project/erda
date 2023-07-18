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

package erda_auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "erda-auth"
)

var (
	_ reverseproxy.RequestFilter = (*ErdaAuth)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type ErdaAuth struct {
	Config *Config
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &ErdaAuth{Config: &cfg}, nil
}

func (f *ErdaAuth) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)

	// Check if this plugin is enabled on this request
	on, err := f.checkIfIsEnabledOnTheRequest(infor)
	if err != nil {
		return reverseproxy.Intercept, err
	}
	if !on {
		return reverseproxy.Continue, nil
	}

	// request erda to check the access permission
	orgId := infor.Header().Get(vars.XAIProxyOrgId)
	userId := infor.Header().Get(vars.XAIProxyUserId)
	access, err := f.request(ctx, orgId, userId)
	if err != nil {
		return reverseproxy.Intercept, err
	}
	if access {
		l.Debugf("the user can not access the org, userId: %s, orgId: %s", userId, orgId)
		http.Error(w, "the user cannot access the organization", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}

	// add authorization
	accessKeyId, err := f.getCredential(ctx, infor)
	if err != nil {
		l.Errorf("failed to First credential, name: %s, platform: %s, err: %v", infor.Header().Get(vars.XErdaAIProxySource), "erda", err)
		http.Error(w, "the erda platform cannot access the AI Service", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}
	infor.Header().Set("Authorization", "Bearer "+accessKeyId)

	return reverseproxy.Continue, nil
}

func (f *ErdaAuth) checkIfIsEnabledOnTheRequest(infor reverseproxy.HttpInfor) (bool, error) {
	for i, item := range f.Config.On {
		if item == nil || item.Key == "" || item.Operator == "" {
			continue
		}
		switch item.Operator {
		case "exist":
			if infor.Header().Get(item.Key) != "" {
				return true, nil
			}
		case "=":
			if infor.Header().Get(item.Key) == string(item.Value) {
				return true, nil
			}
		default:
			return false, errors.Errorf("invalid config: invalid config.on[%d].operator: %s", i, item.Operator)
		}
	}
	return false, nil
}

func (f *ErdaAuth) request(ctx context.Context, orgId, userId string) (bool, error) {
	u := ctx.Value(vars.CtxKeyErdaOpenapi{}).(url.URL)
	u.Path = "/api/permissions/actions/access"
	req := &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: "org",
			ID:   orgId,
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return false, err
	}
	request, err := http.NewRequest(http.MethodPost, u.String(), &buf)
	if err != nil {
		return false, err
	}
	request.Header.Set(httputil.UserHeader, userId)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, err
	}
	defer func() { _ = response.Body.Close() }()
	var resp apistructs.ScopeRoleAccessResponse
	if err = json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return false, err
	}
	return resp.Header.Success && resp.Data.Access, nil
}

func (f *ErdaAuth) getCredential(ctx context.Context, infor reverseproxy.HttpInfor) (string, error) {
	var (
		q          = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO).Q()
		credential models.AIProxyCredentials
		where      = map[string]any{
			"name":     infor.Header().Get(vars.XErdaAIProxySource),
			"platform": "erda",
		}
	)
	if providerName := infor.Header().Get(vars.XAIProxyProvider); providerName != "" {
		providerInstanceId := infor.Header().Get(vars.XAIProxyProviderInstance)
		if providerInstanceId == "" {
			providerInstanceId = "default"
		}
		where["provider"] = providerName
		where["provider_instance_id"] = providerInstanceId
	}
	if err := q.First(&credential, where).Error; err != nil {
		return "", err
	}
	return credential.AccessKeyId, nil
}

type Config struct {
	On []*On
}

type On struct {
	Key      string          `json:"key" yaml:"key"`
	Operator string          `json:"operator" yaml:"operator"`
	Value    json.RawMessage `json:"value" yaml:"value"`
}
