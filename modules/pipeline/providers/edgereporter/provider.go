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

package edgereporter

import (
	"context"
	"reflect"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/providers/edgepipeline_register"
	"github.com/erda-project/erda/modules/pipeline/providers/edgereporter/db"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
)

type config struct {
	CompensatorDuration time.Duration `file:"COMPENSATOR_DURATION" env:"COMPENSATOR_DURATION" default:"2h"`
	OpenAPIPublicURL    string        `file:"OPENAPI_PUBLIC_URL" env:"OPENAPI_PUBLIC_URL"`
	OpenapiToken        string        `file:"DICE_OPENAPI_TOKEN" env:"DICE_OPENAPI_TOKEN"`
}

type provider struct {
	bdl      *bundle.Bundle
	dbClient *db.Client

	Cfg          *config
	Log          logs.Logger
	LW           leaderworker.Interface
	MySQL        mysqlxorm.Interface
	edgeRegister edgepipeline_register.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithAllAvailableClients())
	p.dbClient = &db.Client{Client: dbclient.Client{Engine: p.MySQL.DB()}}
	if p.edgeRegister.IsEdge() {
		p.LW.OnLeader(p.taskReporter)
		p.LW.OnLeader(p.pipelineReporter)
		p.LW.OnLeader(p.compensatorPipelineReporter)
	}
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	token, err := p.edgeRegister.GetAccessToken(apistructs.OAuth2TokenGetRequest{})
	if err != nil {
		p.Log.Errorf("failed to GetAccessToken, err: %v", err)
		return err
	}
	p.Cfg.OpenapiToken = token.AccessToken
	return nil
}

func init() {
	servicehub.Register("edge-reporter", &servicehub.Spec{
		Services:     []string{"edge-reporter"},
		Types:        []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		Dependencies: nil,
		Description:  "pipeline edge reporter",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
