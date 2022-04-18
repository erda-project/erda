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

package edgepipeline_register

import (
	"context"
	"reflect"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
)

type Config struct {
	IsEdge                     bool          `env:"DICE_IS_EDGE" default:"false"`
	ClusterName                string        `env:"DICE_CLUSTER_NAME"`
	PipelineAddr               string        `env:"PIPELINE_ADDR"`
	PipelineHost               string        `env:"PIPELINE_HOST"`
	ClusterDialEndpoint        string        `file:"cluster_dialer_endpoint" desc:"cluster dialer endpoint"`
	ClusterAccessKey           string        `file:"cluster_access_key" desc:"cluster access key, if specified will doesn't start watcher"`
	RetryConnectDialerInterval time.Duration `file:"retry_cluster_hook_interval" default:"1s"`
}

type provider struct {
	Log logs.Logger
	Cfg *Config
	LW  leaderworker.Interface

	bdl *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	p.LW.OnLeader(p.RegisterEdgeToDialer)
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("edgepipeline_register", &servicehub.Spec{
		Services:     []string{"edgepipeline_register"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: nil,
		Description:  "pipeline edgepipeline register agent",
		ConfigFunc:   func() interface{} { return &Config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
