package tagger

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/kubernetes"
	"github.com/erda-project/erda-infra/providers/kubernetes/watcher"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata/pod"
)

var providerName = plugins.WithPrefixProcessor("k8s-tagger")

type config struct {
	Pod pod.Config `file:"pod"`
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
func (p *provider) Process(data model.ObservableData) (model.ObservableData, error) {
	p.addMetadata(data)
	return data, nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {

	to, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return p.initCache(to)
}

// Run this is optional
func (p *provider) Run(ctx context.Context) error {
	podch := p.Kubernetes.WatchPod(ctx, p.Log.Sub("pod-watcher"), watcher.Selector{
		Namespace:     p.Cfg.Pod.WatchSelector.Namespace,
		LabelSelector: p.Cfg.Pod.WatchSelector.LabelSelector,
		FieldSelector: p.Cfg.Pod.WatchSelector.FieldSelector,
	})
	go p.watchPodChange(ctx, podch)
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
