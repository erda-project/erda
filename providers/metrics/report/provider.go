package report

import (
	"net/http"

	"github.com/recallsong/go-utils/logs"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type define struct{}
type ReportMode string

type config struct {
	ReportConfig *ReportConfig
	QueryConfig  *QueryConfig
}

type CollectorConfig struct {
	Addr     string `file:"addr" env:"COLLECTOR_ADDR" default:"localhost:7076"`
	UserName string `file:"username" env:"COLLECTOR_AUTH_USERNAME"`
	Password string `file:"password" env:"COLLECTOR_AUTH_PASSWORD"`
	Retry    int    `file:"retry" env:"TELEMETRY_REPORT_STRICT_RETRY" default:"3"`
}

type ReportConfig struct {
	Mode    string `file:"mode" default:"performance"`
	UdpHost string `file:"udp_host" env:"HOST_IP" default:"localhost"`
	UdpPort string `file:"upd_port" env:"HOST_PORT" default:"7082"`

	Collector *CollectorConfig `file:"collector"`

	BufferSize int `file:"buffer_size" env:"REPORT_BUFFER_SIZE" default:"200"`
}

type QueryConfig struct {
	MonitorAddr string `file:"monitor_addr" env:"MONITOR_ADDR" default:"localhost:7076"`
}

type provider struct {
	Cfg        *config
	Log        logs.Logger
	httpClient *ReportClient
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
	client := &ReportClient{
		CFG: &config{
			ReportConfig: p.Cfg.ReportConfig,
			QueryConfig:  p.Cfg.QueryConfig,
		},
		HttpClient: &http.Client{},
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
