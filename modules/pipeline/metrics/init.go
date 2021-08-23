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
