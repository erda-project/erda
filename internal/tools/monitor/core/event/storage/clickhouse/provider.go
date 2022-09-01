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
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/tools/monitor/core/event"
	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit"
	tablepkg "github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/creator"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

type config struct {
	QueryTimeout time.Duration `file:"query_timeout" default:"1m"`
}

type provider struct {
	Cfg       *config
	Log       logs.Logger
	Creator   creator.Interface   `autowired:"clickhouse.table.creator@event" optional:"true"`
	Loader    loader.Interface    `autowired:"clickhouse.table.loader@event"`
	Retention retention.Interface `autowired:"storage-retention-strategy@event" optional:"true"`

	clickhouse clickhouse.Interface
}

func (p *provider) Init(ctx servicehub.Context) error {
	svc := ctx.Service("clickhouse@event")
	if svc == nil {
		svc = ctx.Service("clickhouse")
	}
	if svc == nil {
		return fmt.Errorf("service clickhouse is required")
	}
	p.clickhouse = svc.(clickhouse.Interface)
	return nil
}

func (p *provider) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	return p.clickhouse.NewWriter(&clickhouse.WriterOptions{
		Encoder: func(data interface{}) (item *clickhouse.WriteItem, err error) {
			eventData := data.(*event.Event)
			item = &clickhouse.WriteItem{
				Data: eventData,
			}
			var (
				wait  <-chan error
				table string
			)

			if p.Retention == nil {
				return nil, errors.Errorf("provider storage-retention-strategy@event is required")
			}

			key := p.Retention.GetConfigKey(eventData.Name, eventData.Tags)
			ttl := p.Retention.GetTTL(key)
			if len(key) > 0 {
				wait, table = p.Creator.Ensure(ctx, eventData.Tags["org_name"], key, tablepkg.FormatTTLToDays(ttl))
			} else {
				wait, table = p.Creator.Ensure(ctx, eventData.Tags["org_name"], "", tablepkg.FormatTTLToDays(ttl))
			}

			if wait != nil {
				select {
				case <-wait:
				case <-ctx.Done():
					return nil, storekit.ErrExitConsume
				}
			}
			item.Table = table
			return item, nil
		},
	}), nil
}

func init() {
	servicehub.Register("event-storage-clickhouse", &servicehub.Spec{
		Services:             []string{"event-storage-clickhouse-reader", "event-storage-clickhouse-writer"},
		Dependencies:         []string{"clickhouse"},
		OptionalDependencies: []string{"clickhouse.table.creator"},
		ConfigFunc:           func() interface{} { return &config{} },
		Creator:              func() servicehub.Provider { return &provider{} },
	})
}
