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

package builder

import (
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
)

type BuilderConfig struct {
	Database         string        `file:"database" default:"monitor"`
	DataType         string        `file:"data_type"`
	TTLCheckInterval time.Duration `file:"ttl_check_interval" default:"12h"`
	SeriesTagKeys    string        `file:"series_tag_keys"`
}

func GetClickHouseInf(ctx servicehub.Context, dt odata.DataType) (clickhouse.Interface, error) {
	svc := ctx.Service("clickhouse@" + strings.ToLower(string(dt)))
	if svc == nil {
		svc = ctx.Service("clickhouse")
	}
	if svc == nil {
		return nil, fmt.Errorf("service clickhouse is required")
	}
	ch, ok := svc.(clickhouse.Interface)
	if !ok {
		return nil, fmt.Errorf("convert svc<%T> failed", svc)
	}
	return ch, nil
}
