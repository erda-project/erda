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
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"

// 	"github.com/erda-project/erda/modules/pipeline/spec"
// 	"terminus.io/dice/telemetry/metrics"
// )

// func init() {
// 	os.Setenv("MONITOR_ADDR", "monitor.default.svc.cluster.local:7096")
// }

// func TestTriggerPipelineTotalCounter(t *testing.T) {
// 	PipelineCounterTotalAdd(spec.Pipeline{
// 		PipelineBase: spec.PipelineBase{
// 			Status:      "success",
// 			ClusterName: "terminus-dev",
// 		},
// 	}, 1)
// }

// func TestPipelineTotalCounter(t *testing.T) {
// 	cfg := metrics.NewQueryConfig()
// 	queryAction, _ := metrics.NewClient().NewQuery(cfg)

// 	req := metrics.CreateQueryMetricRequest("dice_pipeline")
// 	now := time.Now()
// 	start, end := now.AddDate(0, 0, -7), now
// 	req = req.StartFrom(start).EndWith(end).Filter("field", "pipeline_total").Filter("type", "success")
// 	resp, err := queryAction.QueryMetric(req)
// 	if err != nil {
// 		return
// 	}
// 	var v interface{}
// 	assert.NoError(t, json.NewDecoder(bytes.NewBuffer(resp.Body)).Decode(&v))
// 	b, err := json.MarshalIndent(v, "", "  ")
// 	assert.NoError(t, err)
// 	fmt.Println(string(b))
// }

// func TestPipelineGauge(t *testing.T) {
// 	for {
// 		PipelineGaugeProcessingAdd(spec.Pipeline{}, 1)
// 		time.Sleep(time.Second * 5)
// 	}
// }
