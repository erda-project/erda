package settings

import (
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type OpenapiSettings interface {
	GetSessionExpire() time.Duration
}

type config struct {
	SessionExpire time.Duration `file:"session_expire" default:"24h"`
}
type provider struct {
	Cfg *config
}

func (p *provider) GetSessionExpire() time.Duration {
	return p.Cfg.SessionExpire
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	return nil
}

func init() {
	servicehub.Register("openapi-settings", &servicehub.Spec{
		Description: "Openapi global settings",
		Services:    []string{"openapi-settings"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
