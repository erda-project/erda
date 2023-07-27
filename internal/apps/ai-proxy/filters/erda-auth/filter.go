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
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
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

	// check if this filter is enabled on this request
	ok, err := f.checkIfIsEnabledOnTheRequest(infor)
	if err != nil {
		return reverseproxy.Intercept, err
	}
	if !ok {
		l.Debugf("erda-auth is not enabled on the request, on: %+v", f.Config.On)
		return reverseproxy.Continue, nil
	}

	// check if the platform is supported by the ai-proxy
	platformName := infor.Header().Get(vars.XAIProxySource)
	urls := ctx.Value(vars.CtxKeyErdaOpenapi{}).(map[string]*url.URL)
	openapi, ok := urls[platformName]
	if !ok || openapi == nil {
		l.Debugf("erda-auth: ai-proxy dose not support the platform %s", platformName)
		return reverseproxy.Intercept, errors.Errorf("erda-auth: ai-proxy dose not support the platform %s", platformName)
	}

	// check permission for the platform,
	// and add authorization to the request header
	accessKeyId, err := f.getCredential(ctx, infor)
	if err != nil {
		l.Errorf("failed to First credential, name: %s, platform: %s, err: %v", infor.Header().Get(vars.XAIProxySource), "erda", err)
		http.Error(w, "the erda platform cannot access the AI Service", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}
	infor.Header().Set("Authorization", vars.ConcatBearer(accessKeyId))

	// request erda to check the access permission of the user in the org
	orgId := infor.Header().Get(vars.XAIProxyOrgId)
	userId := infor.Header().Get(vars.XAIProxyUserId)
	for _, v := range []*string{
		&orgId,
		&userId,
	} {
		if decoded, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(decoded)
		}
	}
	access, err := f.request(ctx, openapi, orgId, userId)
	if err != nil {
		return reverseproxy.Intercept, err
	}
	if access {
		l.Debugf("the user can not access the org, userId: %s, orgId: %s", userId, orgId)
		http.Error(w, "the user cannot access the organization", http.StatusForbidden)
		return reverseproxy.Intercept, nil
	}

	// set orgId into metadata
	metadata := map[string]any{"orgId": orgId}
	if data, err := json.Marshal(metadata); err == nil {
		infor.Header().Set(vars.XAIProxyMetadata, base64.StdEncoding.EncodeToString(data))
	}

	return reverseproxy.Continue, nil
}

func (f *ErdaAuth) checkIfIsEnabledOnTheRequest(infor reverseproxy.HttpInfor) (bool, error) {
	for i, item := range f.Config.On {
		if item == nil {
			continue
		}
		ok, err := item.On(infor.Header())
		if err != nil {
			return false, errors.Wrapf(err, "invalid config: config.on[%d]", i)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (f *ErdaAuth) request(ctx context.Context, openapi *url.URL, orgId, userId string) (bool, error) {
	openapi.Path = "/api/permissions/actions/access"
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
	request, err := http.NewRequest(http.MethodPost, openapi.String(), &buf)
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
	)
	ok, err := (&credential).Getter(q).Where(
		credential.FieldName().Equal("erda"),
		credential.FieldPlatform().Equal(infor.Header().Get(vars.XAIProxySource)),
		credential.FieldPlatform().NotEqual(""),
		credential.FieldEnabled().Equal(true),
		credential.FieldExpiredAt().MoreThan(time.Now()),
	).Get()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("platform permission denied")
	}
	return credential.AccessKeyID, nil
}

type Config struct {
	On []*common.On `json:"on" yaml:"on"`
}
