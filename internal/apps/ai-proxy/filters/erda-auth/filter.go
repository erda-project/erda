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
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
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

func (f *ErdaAuth) Enable(_ context.Context, req *http.Request) bool {
	for _, item := range f.Config.On {
		if item == nil {
			continue
		}
		if ok, _ := item.On(req.Header); ok {
			return true
		}
	}
	return false
}

func (f *ErdaAuth) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger)

	// check if the platform is supported by the ai-proxy
	platformName := infor.Header().Get(vars.XAIProxySource)
	if platformName == "" {
		http.Error(w, vars.XAIProxySource+" must be set", http.StatusBadRequest)
		return reverseproxy.Intercept, nil
	}
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
	var (
		orgId  = infor.Header().Get(vars.XAIProxyOrgId)
		userId = infor.Header().Get(vars.XAIProxyUserId)
	)
	if orgId == "" {
		http.Error(w, vars.XAIProxyOrgId+" must be set", http.StatusBadRequest)
		return reverseproxy.Intercept, nil
	}
	for _, v := range []*string{
		&orgId,
		&userId,
	} {
		// try to decode base64
		if decoded, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(decoded)
		}
	}

	// set orgId into metadata
	metadata := map[string]any{"orgId": orgId, "userId": userId}
	if data, err := json.Marshal(metadata); err == nil {
		infor.Header().Set(vars.XAIProxyMetadata, base64.StdEncoding.EncodeToString(data))
	}

	return reverseproxy.Continue, nil
}

func (f *ErdaAuth) getOpenapiSessionCookie(infor reverseproxy.HttpInfor) (*http.Cookie, error) {
	cookie, err := infor.Cookie(httputil.CookieNameOpenapiSession)
	if err == nil {
		return cookie, nil
	}
	if !errors.Is(err, http.ErrNoCookie) {
		return nil, err
	}
	openapiSession := infor.Header().Get(vars.XAiProxyErdaOpenapiSession)
	if openapiSession == "" {
		return nil, http.ErrNoCookie
	}
	return &http.Cookie{Name: httputil.CookieNameOpenapiSession, Value: openapiSession}, nil
}

func (f *ErdaAuth) getCredential(ctx context.Context, infor reverseproxy.HttpInfor) (string, error) {
	//var (
	//	q          = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO).Q()
	//	credential models.AIProxyCredentials
	//)
	//ok, err := (&credential).Retriever(q).Where(
	//	credential.FieldName().Equal(infor.Header().Get(vars.XAIProxySource)),
	//	credential.FieldName().NotEqual(""),
	//	credential.FieldPlatform().Equal("erda"),
	//	credential.FieldEnabled().Equal(true),
	//	credential.FieldExpiredAt().MoreThan(time.Now()),
	//).Get()
	//if err != nil {
	//	return "", err
	//}
	//if !ok {
	//	return "", errors.New("platform permission denied")
	//}
	//return credential.AccessKeyID, nil
	return "", nil
}

type Config struct {
	On []*common.On `json:"on" yaml:"on"`
}
