package tagger

import (
	"context"
	"strings"
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
	PodSelector selector    `file:"pod_selector"`
	Matchers    matchersCfg `file:"matchers"`
}

type matchersCfg struct {
	Pod *podMatcherCfg `file:"pod"`
}

type podMatcherCfg struct {
	Filters []filterCfg `file:"filters"`
	Finder  finderCfg   `file:"finder"`
}

type filterCfg struct {
	Key   string `file:"key"`
	Value string `file:"value"`
}

type finderCfg struct {
	NameKey      string `file:"name_key"`
	NamespaceKey string `file:"namespace_key"`
}

type selector struct {
	Namespace     string `file:"namespace"`
	LabelSelector string `file:"label_selector"`
	FieldSelector string `file:"field_selector"`
}

// +provider
type provider struct {
	Cfg        *config
	Log        logs.Logger
	Kubernetes kubernetes.Interface `autowired:"kubernetes"`

	label string
	// cache
	podCache *pod.Cache
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(strings.Join([]string{providerName, p.label}, "@"))
}

func (p *provider) Process(data model.ObservableData) (model.ObservableData, error) {
	switch data.(type) {
	case *model.Metrics:
		return p.processMetrics(data.(*model.Metrics))
	}
	return data, nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.label = ctx.Label()

	to, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return p.initCache(to)
}

// Run this is optional
func (p *provider) Run(ctx context.Context) error {
	podch := p.Kubernetes.WatchPod(ctx, p.Log.Sub("pod-watcher"), watcher.Selector{
		Namespace:     p.Cfg.PodSelector.Namespace,
		LabelSelector: p.Cfg.PodSelector.LabelSelector,
		FieldSelector: p.Cfg.PodSelector.FieldSelector,
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
