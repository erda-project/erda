package stdout

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixExporter("stdout")

type config struct {
}

// +provider
type provider struct {
	Cfg *config
	Log logs.Logger

	label string
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(strings.Join([]string{providerName, p.label}, "@"))
}

func (p *provider) Connect() error {
	return nil
}

func (p *provider) Close() error {
	return nil
}

func (p *provider) Export(od model.ObservableData) error {
	buf, err := json.Marshal(&od)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(buf))
	return nil
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.label = ctx.Label()
	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of erda.oap.collector.exporter.stdout",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
