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
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/config"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/core/pipeline"
	"github.com/erda-project/erda/modules/oap/collector/core/pipeline/metrics"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

type collector struct {
	pipelines []pipeline.Pipeline
	Log       logs.Logger
}

func newCollector(ctx servicehub.Context, cfg *config.Config, log logs.Logger) (*collector, error) {
	c := &collector{
		Log:       log,
		pipelines: make([]pipeline.Pipeline, 0),
	}
	for _, item := range cfg.Pipelines {
		rs, err := findComponents(ctx, item.Receivers)
		if err != nil {
			return nil, err
		}
		ps, err := findComponents(ctx, item.Processors)
		if err != nil {
			return nil, err
		}
		es, err := findComponents(ctx, item.Exporters)
		if err != nil {
			return nil, err
		}

		switch item.DataType {
		case model.MetricDataType:
			pipe := metrics.NewPipeline(log.Sub("MetricsPipeline"))
			err := pipe.InitComponents(rs, ps, es)
			if err != nil {
				return nil, fmt.Errorf("init components err: %w", err)
			}
			c.pipelines = append(c.pipelines, pipe)
		case model.TraceDataType:
		case model.LogDataType:
		default:
			return nil, fmt.Errorf("unsupported data_type: %s", item.DataType)
		}
	}
	return c, nil
}

func (c *collector) start(ctx context.Context) {
	for _, pipe := range c.pipelines {
		go func(pi pipeline.Pipeline) {
			pi.StartStream(ctx)
		}(pipe)
	}
}

func findComponents(ctx servicehub.Context, components []string) ([]model.Component, error) {
	res := make([]model.Component, 0)
	for _, item := range components {
		switch {
		case strings.HasPrefix(item, plugins.PrefixReceiver):
			item = strings.TrimLeft(item, plugins.PrefixReceiver)
		case strings.HasPrefix(item, plugins.PrefixProcessor):
			item = strings.TrimLeft(item, plugins.PrefixProcessor)
		case strings.HasPrefix(item, plugins.PrefixExporter):
			item = strings.TrimLeft(item, plugins.PrefixExporter)
		}

		obj := ctx.Service(item)
		if obj == nil {
			return nil, fmt.Errorf("component %s not found", item)
		}
		com, ok := obj.(model.Component)
		if !ok {
			return nil, fmt.Errorf("%s is not a Component", item)
		}
		res = append(res, com)
	}
	return res, nil
}
