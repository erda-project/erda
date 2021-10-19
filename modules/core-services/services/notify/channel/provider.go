// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package channel

import (
	"github.com/jinzhu/gorm"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/i18n"
	pb "github.com/erda-project/erda-proto-go/core/services/notify/channel/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/services/notify/channel/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg                 *config
	Log                 logs.Logger
	Register            transport.Register
	bdl                 *bundle.Bundle
	notifyChanelService *notifyChannelService
	DB                  *gorm.DB        `autowired:"mysql-client"`
	I18n                i18n.Translator `autowired:"i18n" translator:"cs-i18n"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bdl = bundle.New(bundle.WithScheduler(), bundle.WithCoreServices())
	p.notifyChanelService = &notifyChannelService{
		p:               p,
		NotifyChannelDB: &db.NotifyChannelDB{DB: p.DB},
	}
	if p.Register != nil {
		pb.RegisterNotifyChannelServiceImp(p.Register, p.notifyChanelService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.services.notify.channel.NotifyChanelService" || ctx.Type() == pb.NotifyChannelServiceServerType() || ctx.Type() == pb.NotifyChannelServiceHandlerType():
		return p.notifyChanelService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.services.notify.channel", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
