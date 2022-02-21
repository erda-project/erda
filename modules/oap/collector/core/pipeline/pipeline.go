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
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
)

type Pipeline struct {
	receivers  []model.ReceiverUnit
	processors []model.ProcessorUnit
	exporters  []model.ExporterUnit

	Log logs.Logger
}

func NewPipeline(log logs.Logger) *Pipeline {
	return &Pipeline{Log: log}
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

func (p *Pipeline) rsFromComponent(coms []model.ComponentUnit) ([]model.ReceiverUnit, error) {
	res := make([]model.ReceiverUnit, 0, len(coms))
	for _, com := range coms {
		c, ok := com.Component.(model.Receiver)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.Name, com.Component)
		}
		res = append(res, model.ReceiverUnit{Name: com.Name, Receiver: c})
	}
	return res, nil
}
func (p *Pipeline) prsFromComponent(coms []model.ComponentUnit) ([]model.ProcessorUnit, error) {
	res := make([]model.ProcessorUnit, 0, len(coms))
	for _, com := range coms {
		c, ok := com.Component.(model.Processor)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.Name, com.Component)
		}
		res = append(res, model.ProcessorUnit{Name: com.Name, Processor: c})
	}
	return res, nil
}
func (p *Pipeline) esFromComponent(coms []model.ComponentUnit) ([]model.ExporterUnit, error) {
	res := make([]model.ExporterUnit, 0, len(coms))
	for _, com := range coms {
		c, ok := com.Component.(model.Exporter)
		if !ok {
			return nil, fmt.Errorf("invalid component<%s> type<%T>", com.Name, com.Component)
		}
		res = append(res, model.ExporterUnit{Name: com.Name, Exporter: c})
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
	for {
		select {
		case data := <-out:
			var wg sync.WaitGroup
			wg.Add(len(p.exporters))
			for _, e := range p.exporters {
				go func(exp model.ExporterUnit, od odata.ObservableData) {
					defer wg.Done()
					// TODO. batch
					err := exp.Exporter.Export([]odata.ObservableData{od})
					if err != nil {
						p.Log.Errorf("Exporter<%s> export data error: %s", exp.Name, err)
					}
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
			// TODO. batch
			batch := []odata.ObservableData{data}
			for _, pr := range p.processors {
				tmp, err := pr.Processor.Process(batch...)
				if err != nil {
					p.Log.Errorf("Processor<%s> process data error: %s", pr.Name, err)
					continue
				}
				batch = tmp
			}

			for _, item := range batch {
				// wait forever
				select {
				case out <- item:
				case <-ctx.Done():
					return
				}
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
