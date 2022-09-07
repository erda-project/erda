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

package notify_record

import (
	"github.com/jinzhu/gorm"
	"github.com/recallsong/go-utils/logs"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/kafka"
)

type define struct{}

type config struct {
	Input kafka.ConsumerConfig `file:"input"`
}

func (d *define) Services() []string {
	return []string{"notify-storage"}
}

func (d *define) Dependencies() []string {
	return []string{"mysql"}
}

func (d *define) Summary() string {
	return "notify storage"
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

type provider struct {
	C     *config
	L     logs.Logger
	mysql *gorm.DB
	Kafka kafka.Interface `autowired:"kafkago"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.mysql = ctx.Service("mysql").(mysql.Interface).DB()
	return nil
}

func (p *provider) Start() error {
	err := p.Kafka.NewConsumer(&p.C.Input, p.invoke)
	return err
}
func (p *provider) Close() error {
	logrus.Debug("not support close kafka consumer")
	return nil
}

func init() {
	servicehub.RegisterProvider("notify-storage", &define{})
}
