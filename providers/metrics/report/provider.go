package report

import (
	"net/http"

	"github.com/recallsong/go-utils/logs"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type define struct{}

type config struct {
	Addr     string
	UserName string
	Password string
	Retry    int
}

type provider struct {
	Cfg        *config
	Log        logs.Logger
	httpClient *reportClient
}

func (d *define) Service() []string {
	return []string{"metric-report-client"}
}

func (d *define) Summary() string {
	return "metric-report-client"
}

func (d *define) Description() string {
	return d.Summary()
}

func (d *define) Config() interface{} {
	return &config{}
}

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

func (p *provider) Init(ctx servicehub.Context) error {
	client := &reportClient{
		cfg: &config{
			Addr:     p.Cfg.Addr,
			UserName: p.Cfg.UserName,
			Password: p.Cfg.Password,
			Retry:    p.Cfg.Retry,
		},
		httpClient: &http.Client{},
	}
	p.httpClient = client
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p.httpClient
}

func init() {
	servicehub.RegisterProvider("metric-report-client", &define{})
}
