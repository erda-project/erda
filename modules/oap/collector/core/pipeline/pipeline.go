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
	"github.com/erda-project/erda/modules/oap/collector/core/model"
)

type Pipeline struct {
	receivers  []model.Receiver
	processors []model.Processor
	exporters  []model.Exporter

	Log logs.Logger
}

func NewPipeline(log logs.Logger) *Pipeline {
	return &Pipeline{Log: log}
}

func (p *Pipeline) InitComponents(receivers, processors, exporters []model.Component) error {
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

func (p *Pipeline) rsFromComponent(coms []model.Component) ([]model.Receiver, error) {
	res := make([]model.Receiver, 0, len(coms))
	for _, com := range coms {
		r, ok := com.(model.Receiver)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.ComponentID(), com)
		}
		res = append(res, r)
	}
	return res, nil
}
func (p *Pipeline) prsFromComponent(coms []model.Component) ([]model.Processor, error) {
	res := make([]model.Processor, 0, len(coms))
	for _, com := range coms {
		r, ok := com.(model.Processor)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.ComponentID(), com)
		}
		res = append(res, r)
	}
	return res, nil
}
func (p *Pipeline) esFromComponent(coms []model.Component) ([]model.Exporter, error) {
	res := make([]model.Exporter, 0, len(coms))
	for _, com := range coms {
		r, ok := com.(model.Exporter)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.ComponentID(), com)
		}
		res = append(res, r)
	}
	return res, nil
}

func (p *Pipeline) StartStream(ctx context.Context) {
	out := make(chan model.ObservableData)
	in := make(chan model.ObservableData)
	go p.StartExporters(ctx, out)

	go p.startProcessors(ctx, in, out)

	p.startReceivers(ctx, in)
}

func (p *Pipeline) StartExporters(ctx context.Context, out <-chan model.ObservableData) {
	for {
		select {
		case data := <-out:
			var wg sync.WaitGroup
			wg.Add(len(p.exporters))
			for _, e := range p.exporters {
				go func(exp model.Exporter, od model.ObservableData) {
					defer wg.Done()
					err := exp.Export(od)
					if err != nil {
						p.Log.Errorf("Exporter<%s> export data error: %s", exp.ComponentID(), err)
					}
				}(e, data.Clone())
			}
			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func (p *Pipeline) startProcessors(ctx context.Context, in <-chan model.ObservableData, out chan<- model.ObservableData) {
	for {
		select {
		case data := <-in:
			for _, pr := range p.processors {
				tmp, err := pr.Process(data)
				if err != nil {
					p.Log.Errorf("Processor<%s> process data error: %s", pr.ComponentID(), err)
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

func (p *Pipeline) startReceivers(ctx context.Context, in chan<- model.ObservableData) {
	for _, r := range p.receivers {
		consumer := func(ms model.ObservableData) {
			select {
			case in <- ms:
			case <-ctx.Done():
			}
		}
		r.RegisterConsumer(consumer)
	}
}
