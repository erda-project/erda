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

package tagger

import (
	"context"
	"time"

	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/kubernetes"
	"github.com/erda-project/erda/modules/oap/collector/common/filter"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata/pod"
)

var providerName = plugins.WithPrefixProcessor("k8s-tagger")

type config struct {
	Filter filter.Config `file:"filter"`
	Pod    pod.Config    `file:"pod"`
}

// +provider
type provider struct {
	Cfg        *config
	Log        logs.Logger
	Kubernetes kubernetes.Interface `autowired:"kubernetes"`

	// cache
	podCache *pod.Cache
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

// 1. filter with config filters
// 2. pass tags to handle
func (p *provider) Process(in odata.ObservableData) (odata.ObservableData, error) {
	in.HandleAttributes(func(attr map[string]string) map[string]string {
		if !p.Cfg.Filter.IsTagpass(attr) {
			return attr
		}
		return p.addPodMetadata(attr)
	})

	return in, nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	to, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return p.initCache(to)
}

// Run this is optional
func (p *provider) Run(ctx context.Context) error {
	w, err := p.Kubernetes.Client().CoreV1().Pods(p.Cfg.Pod.WatchSelector.Namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: p.Cfg.Pod.WatchSelector.LabelSelector,
		FieldSelector: p.Cfg.Pod.WatchSelector.FieldSelector,
	})
	if err != nil {
		return err
	}
	go p.watchChange(ctx, w.ResultChan())
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of erda.oap.collector.processor.k8s-tagger",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
