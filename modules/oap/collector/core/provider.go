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

package core

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/config"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/core/pipeline"
)

// +provider
type provider struct {
	Cfg        *config.Config
	Log        logs.Logger
	servicectx servicehub.Context

	pipelines []*pipeline.Pipeline
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.servicectx = ctx
	p.pipelines = make([]*pipeline.Pipeline, 0)
	return nil
}

// Run this is optional
func (p *provider) Run(ctx context.Context) error {
	err := p.initComponents()
	if err != nil {
		return fmt.Errorf("initComponents err: %w", err)
	}

	p.start(ctx)
	return nil
}

func (p *provider) initComponents() error {
	for idx, item := range p.Cfg.Pipelines {
		rs, err := findComponents(p.servicectx, item.Receivers)
		if err != nil {
			return err
		}
		ps, err := findComponents(p.servicectx, item.Processors)
		if err != nil {
			return err
		}
		es, err := findComponents(p.servicectx, item.Exporters)
		if err != nil {
			return err
		}

		pipe := pipeline.NewPipeline(p.Log.Sub(fmt.Sprintf("core-pipeline-%d", idx)), p.Cfg.GlobalConfig)
		err = pipe.InitComponents(rs, ps, es)
		if err != nil {
			return fmt.Errorf("init components err: %w", err)
		}
		p.pipelines = append(p.pipelines, pipe)
	}
	return nil
}

func (p *provider) start(ctx context.Context) {
	for _, pipe := range p.pipelines {
		go func(pi *pipeline.Pipeline) {
			pi.StartStream(ctx)
		}(pipe)
	}
}

func findComponents(ctx servicehub.Context, components []string) ([]model.ComponentUnit, error) {
	res := make([]model.ComponentUnit, 0)
	for _, item := range components {
		obj := ctx.Service(item)
		if obj == nil {
			return nil, fmt.Errorf("component %s not found", item)
		}
		com, ok := obj.(model.Component)
		if !ok {
			return nil, fmt.Errorf("%s is not a Component", item)
		}
		res = append(res, model.ComponentUnit{
			Component: com,
			Name:      item,
		})
	}
	return res, nil
}

func init() {
	servicehub.Register("erda.oap.collector.core", &servicehub.Spec{
		Services:    []string{},
		Description: "core logic for schedule",
		ConfigFunc: func() interface{} {
			return &config.Config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
