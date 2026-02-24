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

package dingtalkworknotice

import (
	"context"
	"encoding/json"
	"reflect"

	gojsonnet "github.com/google/go-jsonnet"

	"github.com/erda-project/erda-infra/base/servicehub"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/jsonnet"
	"github.com/erda-project/erda/internal/core/legacy/services/dingtalk/api/interfaces"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

type Interface interface {
	Send(param *JsonnetParam) (string, error)
}

type provider struct {
	DingtalkApiClient interfaces.DingTalkApiClientFactory `autowired:"dingtalk.api"`
	Identity          userpb.UserServiceServer
	TemplateParser    *jsonnet.Engine
}

type JsonnetParam struct {
	// api request config jsonnet snippet
	Snippet string
	// request jsonnet top level arguments, key: TLA key, value: TLA value
	TLARaw map[string]interface{}
}

type DingTalkConfig struct {
	AgentId   int64
	AppKey    string
	AppSecret string
	Users     []string
	Title     string
	Content   string
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.TemplateParser = &jsonnet.Engine{
		JsonnetVM: gojsonnet.MakeVM(),
	}
	return nil
}

func (p *provider) Send(w *JsonnetParam) (string, error) {
	d, err := p.getDingTalkConfig(w)
	if err != nil {
		return "", err
	}

	client := p.DingtalkApiClient.GetClient(d.AppKey, d.AppSecret, d.AgentId)
	err = client.SendWorkNotice(d.Users, d.Title, d.Content)
	return "", err
}

func (p *provider) getDingTalkConfig(param *JsonnetParam) (*DingTalkConfig, error) {
	// parse content with jsonnet
	b, err := json.Marshal(param.TLARaw)
	if err != nil {
		return nil, err
	}
	jsonStr, err := p.TemplateParser.EvaluateBySnippet(param.Snippet, []jsonnet.TLACodeConfig{
		{
			Key:   "ctx",
			Value: string(b),
		},
	})
	if err != nil {
		return nil, err
	}

	// get API config from json string
	var d DingTalkConfig
	if err := json.Unmarshal([]byte(jsonStr), &d); err != nil {
		return nil, err
	}

	// find users mobile info
	resp, err := p.Identity.FindUsers(
		apis.WithInternalClientContext(context.Background(), discover.SvcDOP),
		&userpb.FindUsersRequest{IDs: d.Users},
	)
	if err != nil {
		return nil, err
	}
	mobiles := make([]string, 0, len(resp.Data))
	for _, u := range resp.Data {
		if u.Phone != "" {
			mobiles = append(mobiles, u.Phone)
		}
	}
	d.Users = mobiles
	return &d, err
}

func init() {
	servicehub.Register("erda.dop.rule.action.dingtalkworknotice", &servicehub.Spec{
		Services: []string{"erda.core.rule.action.dingtalkworknotice"},
		Types:    []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
