package collector

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/oap/collector/authentication"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var providerName = plugins.WithPrefixReceiver("collector")

type config struct {
	MetadataKeyOfTopic string `file:"metadata_key_of_topic"`
	Auth               struct {
		Skip bool `file:"skip"`
	}
}

// +provider
type provider struct {
	Cfg       *config
	Log       logs.Logger
	Router    httpserver.Router        `autowired:"http-router"`
	Validator authentication.Validator `autowired:"erda.oap.collector.authentication.Validator"`

	consumer model.ObservableDataConsumerFunc
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumer = consumer
}

func (p *provider) tokenAuth() interface{} {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator: func(s string, context echo.Context) (bool, error) {
			clusterName := context.Request().Header.Get(apistructs.AuthClusterKeyHeader)
			if clusterName == "" {
				return false, nil
			}

			if p.Validator.Validate(apistructs.CMPClusterScope, clusterName, s) {
				return true, nil
			}

			return false, nil
		},
		Skipper: func(context echo.Context) bool {
			if p.Cfg.Auth.Skip {
				return true
			}
			return false
		},
	})
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	// old
	p.Router.POST("/api/v1/collect/logs/:source", p.collectLogs, p.tokenAuth())
	p.Router.POST("/api/v1/collect/:metric", p.collectMetric, p.tokenAuth())

	return nil
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services:    []string{providerName},
		Description: "here is description of erda.oap.collector.receiver.collector",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
