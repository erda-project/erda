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

package clickhouse

import (
	"context"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/core/monitor/storekit/clickhouse/table/creator"
)

type config struct {
}

type provider struct {
	Cfg     *config
	Log     logs.Logger
	Creator creator.Interface `autowired:"clickhouse.table.creator"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	// test create table
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	wait, aliasName := p.Creator.Ensure(context, "test", "123")
	err := <-wait
	p.Log.Info(aliasName, err)
	return err
}

func init() {
	servicehub.Register("log-storage-clickhouse", &servicehub.Spec{
		Services:     []string{"log-storage-clickhouse-reader", "log-storage-clickhouse-writer"},
		Dependencies: []string{"clickhouse"},
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
