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

package autotest_cookie_keep_after

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/pkg/apitestsv2"
)

const taskType = "api-test"
const metaKeySetCookie = "api_set_cookie"
const AutotestApiGlobalConfig = "AUTOTEST_API_GLOBAL_CONFIG"
const ReportTypeAutotestSetCookie = "autotest_set_cookie"

type Plugin struct {
	aoptypes.TaskBaseTunePoint
}

func (p *Plugin) Name() string {
	return "autotest_cookie_keep_after"
}

func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {
	// task not api-test type return
	if ctx.SDK.Task.Type != taskType {
		return nil
	}

	// task result metafile not have set_cookie return
	metadata := ctx.SDK.Task.Result.Metadata
	if metadata == nil {
		return nil
	}
	var setCookieJSON string
	for _, field := range metadata {
		if field.Name == metaKeySetCookie {
			setCookieJSON = field.Value
			break
		}
	}
	if setCookieJSON == "" {
		return nil
	}

	if ctx.SDK.Pipeline.Snapshot.Secrets == nil {
		return nil
	}

	// report cookieJar
	_, err := ctx.SDK.Report.Create(apistructs.PipelineReportCreateRequest{
		PipelineID: ctx.SDK.Pipeline.ID,
		Type:       ReportTypeAutotestSetCookie,
		Meta: map[string]interface{}{
			apitestsv2.HeaderSetCookie: setCookieJSON,
		},
	})
	return err
}

func New() *Plugin {
	var p Plugin
	return &p
}
