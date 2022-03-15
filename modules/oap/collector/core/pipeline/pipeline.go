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

package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/modules/oap/collector/common"
	"github.com/erda-project/erda/modules/oap/collector/core/config"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
)

type Pipeline struct {
	cfg        config.GlobalConfig
	receivers  []*model.RuntimeReceiver
	processors []*model.RuntimeProcessor
	exporters  []*model.RuntimeExporter

	Log logs.Logger
}

func NewPipeline(log logs.Logger, cfg config.GlobalConfig) *Pipeline {
	return &Pipeline{Log: log, cfg: cfg}
}

func (p *Pipeline) InitComponents(receivers, processors, exporters []model.ComponentUnit) error {
	rs, err := p.rsFromComponent(receivers)
	if err != nil {
		return err
	}
	prs, err := p.prsFromComponent(processors)
	if err != nil {
		return err
	}
	es, err := p.esFromComponent(exporters)
	if err != nil {
		return err
	}

	p.receivers = rs
	p.processors = prs
	p.exporters = es
	return nil
}

func (p *Pipeline) rsFromComponent(coms []model.ComponentUnit) ([]*model.RuntimeReceiver, error) {
	res := make([]*model.RuntimeReceiver, 0, len(coms))
	for _, com := range coms {
		c, ok := com.Component.(model.Receiver)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.Name, com.Component)
		}
		res = append(res, &model.RuntimeReceiver{Name: com.Name, Receiver: c, Filter: com.Filter})
	}
	return res, nil
}

func (p *Pipeline) prsFromComponent(coms []model.ComponentUnit) ([]*model.RuntimeProcessor, error) {
	res := make([]*model.RuntimeProcessor, 0, len(coms))
	for _, com := range coms {
		c, ok := com.Component.(model.Processor)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.Name, com.Component)
		}
		res = append(res, &model.RuntimeProcessor{Name: com.Name, Processor: c, Filter: com.Filter})
	}
	return res, nil
}
func (p *Pipeline) esFromComponent(coms []model.ComponentUnit) ([]*model.RuntimeExporter, error) {
	res := make([]*model.RuntimeExporter, 0, len(coms))
	for _, com := range coms {
		c, ok := com.Component.(model.Exporter)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.Name, com.Component)
		}
		buffer := odata.NewBuffer(p.cfg.BatchLimit)
		timer := common.NewRunningTimer(p.cfg.FlushInterval, p.cfg.FlushJitter)
		res = append(res, &model.RuntimeExporter{
			Name:     com.Name,
			Logger:   p.Log.Sub("exporter-" + com.Name),
			Exporter: c,
			Filter:   com.Filter,
			Timer:    timer,
			Buffer:   buffer,
		})
	}
	return res, nil
}

func (p *Pipeline) StartStream(ctx context.Context) {
	out := make(chan odata.ObservableData, 10)
	in := make(chan odata.ObservableData, 10)
	go p.StartExporters(ctx, out)

	go p.startProcessors(ctx, in, out)

	p.startReceivers(ctx, in)
}

func (p *Pipeline) StartExporters(ctx context.Context, out <-chan odata.ObservableData) {
	for _, e := range p.exporters {
		go func(exp *model.RuntimeExporter) {
			go exp.Start(ctx)
		}(e)
	}

	for {
		select {
		case data := <-out:
			var wg sync.WaitGroup
			wg.Add(len(p.exporters))
			for _, e := range p.exporters {
				go func(exp *model.RuntimeExporter, od odata.ObservableData) {
					defer wg.Done()
					exp.Add(od)
				}(e, data.Clone())
			}
			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func (p *Pipeline) startProcessors(ctx context.Context, in <-chan odata.ObservableData, out chan<- odata.ObservableData) {
	for _, r := range p.processors {
		rp, ok := r.Processor.(model.RunningProcessor)
		if !ok {
			continue
		}
		rp.StartProcessor(newConsumer(ctx, out))
	}

	for {
		select {
		case data := <-in:
			for _, pr := range p.processors {
				if !pr.Filter.Selected(data) {
					continue
				}
				tmp, err := pr.Processor.Process(data)
				if err != nil {
					p.Log.Errorf("Processor<%s> process data error: %s", pr.Name, err)
					continue
				}
				data = tmp
			}

			// wait forever
			select {
			case out <- data:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *Pipeline) startReceivers(ctx context.Context, out chan<- odata.ObservableData) {
	for _, r := range p.receivers {
		r.Receiver.RegisterConsumer(newConsumer(ctx, out))
	}
}

func newConsumer(ctx context.Context, out chan<- odata.ObservableData) model.ObservableDataConsumerFunc {
	return func(od odata.ObservableData) {
		select {
		case out <- od:

		case <-ctx.Done():
			return
		}
	}
}
