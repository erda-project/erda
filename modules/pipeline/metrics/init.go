// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package metrics

// import (
// 	"github.com/sirupsen/logrus"

// 	"terminus.io/dice/telemetry/metrics"
// )

// // disableMetrics 是否禁用 metric 相关操作
// var disableMetrics bool

// var bulkClient *metrics.BulkAction

// func Initialize() {
// 	client := metrics.NewClient()
// 	bulk, err := client.NewBulk()
// 	if err != nil {
// 		logrus.Errorf("[alert] failed to init event bulk client, disable metric report, err: %v", err)
// 		disableMetrics = true
// 		return
// 	}
// 	bulkClient = bulk
// 	logrus.Info("metrics enabled")
// }
