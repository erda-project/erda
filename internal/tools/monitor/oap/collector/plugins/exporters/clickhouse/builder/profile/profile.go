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

package profile

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/pyroscope-io/pyroscope/pkg/service"
	"github.com/pyroscope-io/pyroscope/pkg/storage/dict"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	mysql "github.com/erda-project/erda-infra/providers/mysql/v2"
	"github.com/erda-project/erda/internal/tools/monitor/core/profile"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
)

const (
	treeTable    = "pyroscope.trees"
	segmentTable = "pyroscope.segments"
	dictTable    = "pyroscope.dicts"
)

const (
	treePrefix    = "t:"
	segmentPrefix = "s:"
	dictPrefix    = "d:"
)

type Builder struct {
	logger         logs.Logger
	client         clickhouse.Conn
	cfg            *builder.BuilderConfig
	appMetadataSvc *service.ApplicationMetadataCacheService

	DB mysql.Interface
}

func NewBuilder(ctx servicehub.Context, logger logs.Logger, cfg *builder.BuilderConfig) (*Builder, error) {
	bu := &Builder{
		cfg:    cfg,
		logger: logger,
	}

	ch, err := builder.GetClickHouseInf(ctx, odata.ProfileType)
	if err != nil {
		return nil, fmt.Errorf("failed to get clickhouse interface: %w", err)
	}
	bu.client = ch.Client()
	db, err := GetMysqlInf(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mysql interface: %w", err)
	}
	bu.DB = db
	bu.appMetadataSvc = service.NewApplicationMetadataCacheService(
		service.ApplicationMetadataCacheServiceConfig{
			Size: 8192,
			TTL:  15 * time.Minute,
		}, service.NewApplicationMetadataService(bu.DB.DB()))

	return bu, nil
}

func (bu *Builder) BuildBatch(ctx context.Context, sourceBatch interface{}) ([]driver.Batch, error) {
	items, ok := sourceBatch.([]*profile.Output)
	if !ok {
		return nil, fmt.Errorf("soureBatch<%T> must be []*profile.Output", sourceBatch)
	}
	// nolint
	batches, err := bu.buildBatches(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("failed buildBatches: %w", err)
	}
	return batches, nil
}

func (bu *Builder) buildBatches(ctx context.Context, items []*profile.Output) ([]driver.Batch, error) {
	treeBatch, err := bu.client.PrepareBatch(ctx, "INSERT INTO "+treeTable, driver.WithReleaseConnection())
	if err != nil {
		return nil, fmt.Errorf("failed to get tree table: %v", err)
	}
	segmentBatch, err := bu.client.PrepareBatch(ctx, "INSERT INTO "+segmentTable, driver.WithReleaseConnection())
	if err != nil {
		return nil, fmt.Errorf("failed to get segment table: %v", err)
	}
	dictBatch, err := bu.client.PrepareBatch(ctx, "INSERT INTO "+dictTable, driver.WithReleaseConnection())
	if err != nil {
		return nil, fmt.Errorf("failed to get dict table: %v", err)
	}
	for _, item := range items {
		for _, data := range item.Profiles() {
			bu.appMetadataSvc.CreateOrUpdate(context.Background(), *data.App)
			treeVal, dictVal, err := bu.splitTree(data.Tree)
			if err != nil {
				return nil, fmt.Errorf("failed to split tree: %v", err)
			}
			segmentVal, err := bu.getSegmentVal(data.Segment)
			if err != nil {
				return nil, fmt.Errorf("failed to get segment val: %v", err)
			}
			dictKey := segment.FromTreeToDictKey(data.TreeKey)
			treeBatch.AppendStruct(&profile.TableGeneral{
				K:         treePrefix + data.TreeKey,
				V:         treeVal,
				Timestamp: *data.CollectTime,
			})
			segmentBatch.AppendStruct(&profile.TableGeneral{
				K:         segmentPrefix + data.SegmentKey,
				V:         segmentVal,
				Timestamp: *data.CollectTime,
			})
			dictBatch.AppendStruct(&profile.TableGeneral{
				K:         dictPrefix + dictKey,
				V:         dictVal,
				Timestamp: *data.CollectTime,
			})
		}
	}
	return []driver.Batch{treeBatch, segmentBatch, dictBatch}, nil
}

func (bu *Builder) getSegmentVal(seg *segment.Segment) (string, error) {
	segWriter := new(bytes.Buffer)
	err := seg.Serialize(segWriter)
	if err != nil {
		return "", err
	}
	return segWriter.String(), nil
}

func (bu *Builder) splitTree(t *tree.Tree) (string, string, error) {
	treeWriter := new(bytes.Buffer)
	d := dict.New()
	err := t.SerializeTruncate(d, 8192, treeWriter)
	if err != nil {
		return "", "", err
	}

	dictWriter := new(bytes.Buffer)
	err = d.Serialize(dictWriter)
	if err != nil {
		return "", "", err
	}
	return treeWriter.String(), dictWriter.String(), nil
}

func GetMysqlInf(ctx servicehub.Context) (mysql.Interface, error) {
	svc := ctx.Service("mysql-gorm.v2")
	if svc == nil {
		return nil, fmt.Errorf("service gorm is required")
	}
	db, ok := svc.(mysql.Interface)
	if !ok {
		return nil, fmt.Errorf("convert svc<%T> failed", svc)
	}
	return db, nil
}
