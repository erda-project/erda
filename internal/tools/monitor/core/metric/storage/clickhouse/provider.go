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

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"

	"github.com/erda-project/erda/internal/tools/monitor/core/storekit"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
)

type (
	config struct {
		QueryTimeout            time.Duration `file:"query_timeout" default:"1m"`
		QueryMaxThreads         int           `file:"query_max_threads" default:"0"`
		QueryMaxMemory          int64         `file:"query_max_memory" default:"0"`
		RuntimeSettings         []string      `file:"runtime_settings"`
		ExternalDefaultDatabase string        `file:"external_default_database" default:"monitor"`
		ExternalDefaultTable    string        `file:"external_default_table" default:"external_metrics_all"`
	}
	provider struct {
		Cfg    *config
		Log    logs.Logger
		Loader loader.Interface `autowired:"clickhouse.table.loader@metric"`

		clickhouse clickhouse.Interface
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	svc := ctx.Service("clickhouse@metric")
	if svc == nil {
		svc = ctx.Service("clickhouse")
	}
	if svc == nil {
		return fmt.Errorf("service clickhouse is required")
	}
	p.clickhouse = svc.(clickhouse.Interface)

	return nil
}
func init() {
	servicehub.Register("metric-storage-clickhouse", &servicehub.Spec{
		Services:   []string{"metric-storage-clickhouse"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}

func (p *provider) Select(metrics []string) bool {
	return true
}
func (p *provider) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	return nil, nil
}
