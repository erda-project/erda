package stdout

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/extensions/loghub/exporter"
	"github.com/recallsong/go-utils/reflectx"
)

type define struct{}

func (d *define) Service() []string      { return []string{"logs-exporter-stdout"} }
func (d *define) Dependencies() []string { return []string{"logs-exporter-base"} }
func (d *define) Summary() string        { return "logs export to stdout" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type provider struct {
	exp exporter.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.exp = ctx.Service("logs-exporter-base").(exporter.Interface)
	return nil
}

func (p *provider) Start() error {
	return p.exp.NewConsumer(p.newOutput)
}

func (p *provider) Close() error { return nil }

func (p *provider) newOutput(i int) (exporter.Output, error) {
	return p, nil
}

func (p *provider) Write(key string, data []byte) error {
	fmt.Println(key, reflectx.BytesToString(data))
	return nil
}

func init() {
	servicehub.RegisterProvider("logs-exporter-stdout", &define{})
}
