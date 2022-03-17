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

package eventbox

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	eventpb "github.com/erda-project/erda-proto-go/core/messenger/eventbox/pb"
	"github.com/erda-project/erda-proto-go/core/messenger/notify/pb"
	"github.com/erda-project/erda/modules/core-services/services/dingtalk/api/interfaces"
	"github.com/erda-project/erda/modules/messenger/eventbox/input/http"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/sirupsen/logrus"
)

type config struct{}

type provider struct {
	DingtalkApiClient interfaces.DingTalkApiClientFactory `autowired:"dingtalk.api"`
	Messenger         pb.NotifyServiceServer

	Register        transport.Register `autowired:"service-register" optional:"true"`
	C               *config
	eventBoxService *eventBoxService
}

func (p *provider) Run(ctx context.Context) error {
	//return Initialize(p.DingtalkApiClient, p.Messenger)
	p.eventBoxService.HttpI.InputName = "pjytest"
	return Initialize(p.DingtalkApiClient, p.Messenger, p.eventBoxService.HttpI)
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.eventBoxService = &eventBoxService{}
	httpi, err := http.New()
	if err != nil {
		logrus.Error("HttpInput init is failed err is ", err)
		return err
	}
	p.eventBoxService.HttpI = httpi
	if p.Register != nil {
		type EventBoxService = eventpb.EventBoxServiceServer
		eventpb.RegisterEventBoxServiceImp(p.Register, p.eventBoxService, apis.Options())
	}
	return nil
}

func init() {
	servicehub.Register("eventbox", &servicehub.Spec{
		Services:             []string{"eventbox"},
		Creator:              func() servicehub.Provider { return &provider{} },
		OptionalDependencies: []string{"service-register"},
		Description:          "",
	})
}
