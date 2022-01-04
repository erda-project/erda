package modifier

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/common/filter"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixProcessor("modifier")

type config struct {
	NameOverride string        `file:"name_override"`
	Filter       filter.Config `file:"filter"`
	Rules        []modifierCfg `file:"rules"`
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

func (p *provider) Process(data model.ObservableData) (model.ObservableData, error) {
	data.RangeTagsFunc(func(tags map[string]string) map[string]string {
		if filter.IsInclude(p.Cfg.Filter, tags) {
			tags = p.modify(tags)
		}
		return tags
	})
	if p.Cfg.NameOverride != "" {
		data.RangeNameFunc(func(name string) string {
			return p.Cfg.NameOverride
		})
	}
	return data, nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},

		Description: "here is description of erda.oap.collector.processor.modifier",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
