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
		WriteTimeout time.Duration `file:"write_timeout" default:"1m"`
		IndexType    string        `file:"index_type" default:"metric"`
		DummyIndex   string        `file:"dummy_index"`
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

func (p *provider) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	return nil, nil
}
