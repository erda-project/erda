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

package log

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/tools/monitor/core/log"
	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	tablepkg "github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/creator"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
)

const (
	chTableCreator   = "clickhouse.table.creator@log"
	chTableRetention = "storage-retention-strategy@log"
)

type Builder struct {
	logger    logs.Logger
	client    clickhouse.Conn
	Creator   creator.Interface
	Retention retention.Interface
	cfg       *builder.BuilderConfig
}

func NewBuilder(ctx servicehub.Context, logger logs.Logger, cfg *builder.BuilderConfig) (*Builder, error) {
	bu := &Builder{
		cfg:    cfg,
		logger: logger,
	}
	ch, err := builder.GetClickHouseInf(ctx, odata.LogType)
	if err != nil {
		return nil, fmt.Errorf("get clickhouse interface: %w", err)
	}
	bu.client = ch.Client()

	if svc, ok := ctx.Service(chTableCreator).(creator.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableCreator)
	} else {
		bu.Creator = svc
	}

	if svc, ok := ctx.Service(chTableRetention).(retention.Interface); !ok {
		return nil, fmt.Errorf("service %q must existed", chTableRetention)
	} else {
		bu.Retention = svc
	}

	return bu, nil
}

func (bu *Builder) BuildBatch(ctx context.Context, sourceBatch interface{}) ([]driver.Batch, error) {
	items, ok := sourceBatch.([]*log.Log)
	if !ok {
		return nil, fmt.Errorf("sourceBatch<%T> must be []*log.LabeldLog", sourceBatch)
	}

	batches, err := bu.buildBatches(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("failed buildBatches: %w", err)
	}
	return batches, nil
}

func (bu *Builder) buildBatches(ctx context.Context, items []*log.Log) ([]driver.Batch, error) {
	tableBatch := make(map[string]driver.Batch)
	for _, item := range items {
		bu.fillLogInfo(item)
		table, err := bu.getOrCreateTenantTable(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("cannot get tenant table: %w", err)
		}
		if _, ok := tableBatch[table]; !ok {
			batch, err := bu.client.PrepareBatch(ctx, "INSERT INTO "+table)
			if err != nil {
				return nil, err
			}
			tableBatch[table] = batch
		}
		batch := tableBatch[table]
		batch.AppendStruct(item)
	}
	batches := make([]driver.Batch, 0, len(tableBatch))
	for _, batch := range tableBatch {
		batches = append(batches, batch)
	}
	return batches, nil
}

func (bu *Builder) getOrCreateTenantTable(ctx context.Context, data *log.Log) (string, error) {
	key := bu.Retention.GetConfigKey(data.Source, data.Tags)
	ttl := bu.Retention.GetTTL(key)
	var (
		wait  <-chan error
		table string
	)
	if len(key) > 0 {
		wait, table = bu.Creator.Ensure(ctx, data.Tags["dice_org_name"], key, tablepkg.FormatTTLToDays(ttl))
	} else {
		wait, table = bu.Creator.Ensure(ctx, data.Tags["dice_org_name"], "", tablepkg.FormatTTLToDays(ttl))
	}
	if wait != nil {
		select {
		case <-wait:
		case <-ctx.Done():
			return table, nil
		}
	}
	return table, nil
}

func (bu *Builder) fillLogInfo(logData *log.Log) {
	id := logData.ID
	if len(id) > 12 {
		id = id[:12]
	}
	logData.UniqId = strconv.FormatInt(logData.Timestamp, 36) + "-" + id
	logData.OrgName = logData.Tags["dice_org_name"]
	tenantId := logData.Tags["monitor_log_key"]
	if len(tenantId) == 0 {
		tenantId = logData.Tags["msp_env_id"]
	}
	logData.WriteTimestamp = time.Unix(0, logData.Timestamp)
	logData.TenantId = tenantId
}
