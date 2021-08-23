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

package autotest_cookie_keep_before

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/task/plugins/autotest_cookie_keep_after"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/pkg/apitestsv2"
)

const taskType = "api-test"

type Plugin struct {
	aoptypes.TaskBaseTunePoint
}

func (p *Plugin) Name() string {
	return "autotest_cookie_keep_before"
}

func (p *Plugin) Handle(ctx *aoptypes.TuneContext) error {
	// task not api-test type return
	if ctx.SDK.Task.Type != taskType {
		return nil
	}

	// search from report
	// depends on creation time in reverse order
	// will only fetch the latest one
	reportSets, err := ctx.SDK.Report.GetPipelineReportSet(ctx.SDK.Pipeline.ID, autotest_cookie_keep_after.ReportTypeAutotestSetCookie)
	if err != nil {
		rlog.TErrorf(ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, "failed to get pipeline reports, err: %", err)
		return err
	}
	var setCookieJSON string
	for _, v := range reportSets.Reports {
		if v.Meta == nil || v.Meta[apitestsv2.HeaderSetCookie] == nil {
			continue
		}
		setCookieJSON = v.Meta[apitestsv2.HeaderSetCookie].(string)
		break
	}
	rlog.TDebugf(ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, "setCookieJSON: %s", setCookieJSON)
	if setCookieJSON == "" {
		return nil
	}
	// parse Set-Cookie-JSON to Cookie
	var setCookies []string
	if err := json.Unmarshal([]byte(setCookieJSON), &setCookies); err != nil {
		return fmt.Errorf("failed to parse Set-Cookie: %s, err: %v", setCookieJSON, err)
	}
	if len(setCookies) == 0 {
		return nil
	}
	setCookie := setCookies[0]

	logrus.Infof("pipelineID: %d, taskID: %d, autotest keep cookie: %v", ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, setCookie)
	// if autoTestAPIConfig is empty
	// means not use config to run, also need to keep cookie
	var config apistructs.AutoTestAPIConfig
	if configStr, ok := ctx.SDK.Task.Extra.PrivateEnvs[autotest_cookie_keep_after.AutotestApiGlobalConfig]; ok {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			rlog.TErrorf(ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, "failed to unmarshal AUTOTEST_API_GLOBAL_CONFIG, err: %v", err)
			return err
		}
	}
	if config.Header == nil {
		config.Header = map[string]string{}
	}
	config.Header[apitestsv2.HeaderCookie] = setCookie
	configJson, err := json.Marshal(&config)
	if err != nil {
		rlog.TErrorf(ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, "failed to marshal AUTOTEST_API_GLOBAL_CONFIG, err: %v", err)
		return err
	}
	ctx.SDK.Task.Extra.PrivateEnvs[autotest_cookie_keep_after.AutotestApiGlobalConfig] = string(configJson)

	err = ctx.SDK.DBClient.UpdatePipelineTaskExtra(ctx.SDK.Task.ID, ctx.SDK.Task.Extra)
	if err != nil {
		rlog.TErrorf(ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, "failed to update task extra, err: %v", err)
		return err
	}
	rlog.TDebugf(ctx.SDK.Pipeline.ID, ctx.SDK.Task.ID, "AUTOTEST_API_GLOBAL_CONFIG updated")
	return nil
}

func New() *Plugin {
	var p Plugin
	return &p
}
