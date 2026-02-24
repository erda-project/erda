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

package eventsync

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	useroauthpb "github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/user/impl/uc/eventsync/dao"
	"github.com/erda-project/erda/internal/core/user/impl/uc/eventsync/ucclient"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type provider struct {
	Cfg *config
	Log logs.Logger

	UserOAuthSvc useroauthpb.UserOAuthServiceServer

	client *httpclient.HTTPClient
	bdl    *bundle.Bundle
	db     *dao.DBClient
	syncer *Syncer
}

func (p *provider) Init(_ servicehub.Context) error {
	p.client = httpclient.New()
	p.bdl = bundle.New(bundle.WithErdaServer())

	db, err := dao.Open()
	if err != nil {
		return err
	}
	p.db = db

	uc := ucclient.NewUCClient(p.Cfg.Host, p.client, p.UserOAuthSvc)
	p.syncer = NewSyncer(
		WithDBClient(db),
		WithUCClient(uc),
		WithBundle(p.bdl),
		WithConfig(p.Cfg),
		WithLogger(p.Log),
	)

	return nil
}

func WithLogger(log logs.Logger) Option {
	return func(syncer *Syncer) {
		syncer.log = log
	}
}

func (p *provider) Start() error {
	if p.syncer == nil {
		return nil
	}
	p.syncer.Start()
	return nil
}

func (p *provider) Close() error {
	if p.syncer != nil {
		p.syncer.Close()
	}
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func init() {
	servicehub.Register("erda.core.user.uc.event-sync", &servicehub.Spec{
		Services:   []string{"erda.core.user.uc.event-sync"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
