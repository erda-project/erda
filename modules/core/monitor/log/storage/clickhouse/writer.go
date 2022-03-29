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
	"strconv"

	"github.com/erda-project/erda/modules/core/monitor/log"

	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
)

func (p *provider) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	return p.clickhouse.NewWriter(&clickhouse.WriterOptions{
		Encoder: func(data interface{}) (item *clickhouse.WriteItem, err error) {
			logData := data.(*log.LabeledLog)
			item = &clickhouse.WriteItem{
				Data: &logData.Log,
			}
			var wait <-chan error
			var table string

			if p.Retention == nil {
				wait, table = p.Creator.Ensure(ctx, logData.Tags["dice_org_name"], "")
			} else {
				key := p.Retention.GetConfigKey(logData.Source, logData.Tags)
				if len(key) > 0 {
					wait, table = p.Creator.Ensure(ctx, logData.Tags["dice_org_name"], key)
				} else {
					wait, table = p.Creator.Ensure(ctx, logData.Tags["dice_org_name"], "")
				}
			}
			if wait != nil {
				select {
				case <-wait:
				case <-ctx.Done():
					return nil, storekit.ErrExitConsume
				}
			}

			id := logData.ID
			if len(id) > 12 {
				id = id[:12]
			}
			logData.Log.UniqId = id + strconv.FormatInt(logData.Timestamp, 36) + "-" + strconv.FormatInt(logData.Offset, 36)
			logData.OrgName = logData.Tags["dice_org_name"]
			item.Table = table
			return item, nil
		},
	}), nil
}
