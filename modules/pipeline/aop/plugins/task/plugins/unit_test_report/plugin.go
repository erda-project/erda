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

package unit_test_report

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
)

const taskType = "unit-test"
const actionTypeUnitTest = "unit-test"

type Plugin struct {
	aoptypes.TaskBaseTunePoint
}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Name() string { return "unit-test-report" }
func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {

	if ctx.SDK.Task.Type != taskType {
		return nil
	}

	metadata := ctx.SDK.Task.Result.Metadata
	if metadata == nil {
		return nil
	}

	var meta = map[string]interface{}{}
	for _, v := range metadata {
		var err error
		switch v.Name {
		case "results":
			var results apistructs.TestResults
			err = json.Unmarshal([]byte(v.Value), &results)
			meta["results"] = results
		case "totals":
			var totals apistructs.TestTotals
			err = json.Unmarshal([]byte(v.Value), &totals)
			meta["totals"] = totals
		case "suites":
			var suites []apistructs.TestSuite
			err = json.Unmarshal([]byte(v.Value), &suites)
			meta["suites"] = suites
		}
		if err != nil {
			return fmt.Errorf("unmarshal unit-test report error: %v", err)
		}
	}

	meta["taskId"] = ctx.SDK.Task.ID

	_, err := ctx.SDK.Report.Create(apistructs.PipelineReportCreateRequest{
		PipelineID: ctx.SDK.Pipeline.ID,
		Type:       actionTypeUnitTest,
		Meta:       meta,
	})
	if err != nil {
		return err
	}

	return nil
}
