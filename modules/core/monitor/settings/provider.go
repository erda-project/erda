// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package settings

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/settings/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type config struct{}

// +provider
type provider struct {
	Cfg             *config
	Log             logs.Logger
	Register        transport.Register `autowired:"service-register" optional:"true"`
	DB              *gorm.DB           `autowired:"mysql-client"`
	Trans           i18n.Translator    `autowired:"i18n" translator:"settings"`
	settingsService *settingsService
}

func (p *provider) Init(ctx servicehub.Context) error {
	bundle := bundle.New(
		bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))),
		bundle.WithCoreServices(),
	)

	p.settingsService = &settingsService{p: p, db: p.DB, bundle: bundle, t: p.Trans}
	p.settingsService.initConfigMap()
	if p.Register != nil {
		pb.RegisterSettingsServiceImp(p.Register, p.settingsService, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithEncoder(func(rw http.ResponseWriter, r *http.Request, data interface{}) error {
					if resp, ok := data.(*apis.Response); ok && resp != nil {
						if data, ok := resp.Data.(map[string]*pb.ConfigGroups); ok {
							m := make(map[string]interface{})
							for key, val := range data {
								if val != nil {
									m[key] = val.Groups
								}
							}
							resp.Data = m
						}
					}
					return encoding.EncodeResponse(rw, r, data)
				}),
				transhttp.WithDecoder(func(r *http.Request, out interface{}) error {
					if body, ok := out.(*map[string]*pb.ConfigGroups); ok {
						recv := make(map[string][]*pb.ConfigGroup)
						err := encoding.DecodeRequest(r, &recv)
						if err != nil {
							return err
						}
						m := make(map[string]*pb.ConfigGroups)
						for key, groups := range recv {
							m[key] = &pb.ConfigGroups{
								Groups: groups,
							}
						}
						*body = m
						return nil
					} else if body, ok := out.(*[]*pb.MonitorConfig); ok {
						var recv []*pb.MonitorConfig
						err := encoding.DecodeRequest(r, &recv)
						if err != nil {
							return err
						}
						*body = recv
						return nil
					}
					return encoding.DecodeRequest(r, out)
				}),
			),
		)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.monitor.settings.SettingsService" || ctx.Type() == pb.SettingsServiceServerType() || ctx.Type() == pb.SettingsServiceHandlerType():
		return p.settingsService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.monitor.settings", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		ConfigFunc:           func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
