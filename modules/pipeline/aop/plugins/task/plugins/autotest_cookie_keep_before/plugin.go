package autotest_cookie_keep_before

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/aop/plugins/task/plugins/autotest_cookie_keep_after"
)

const taskType = "api-test"

type Plugin struct {
	aoptypes.TaskBaseTunePoint
}

func (p *Plugin) Name() string {
	return "autotest_cookie_keep_before"
}

func (p *Plugin) Handle(ctx aoptypes.TuneContext) error {
	// task not api-test type return
	if ctx.SDK.Task.Type != taskType {
		return nil
	}

	// search from report
	// depends on creation time in reverse order
	// will only fetch the latest one
	reportSets, err := ctx.SDK.Report.GetPipelineReportSet(ctx.SDK.Pipeline.ID, autotest_cookie_keep_after.ReportTypeAutotestSetCookie)
	if err != nil {
		return err
	}
	var cookieJar string
	for _, v := range reportSets.Reports {
		if v.Meta == nil || v.Meta[autotest_cookie_keep_after.CookieJar] == nil {
			continue
		}
		cookieJar = v.Meta[autotest_cookie_keep_after.CookieJar].(string)
		break
	}

	if len(cookieJar) <= 0 {
		return nil
	}

	logrus.Infof("autotest keep cookie： %v", cookieJar)
	// if autoTestAPIConfig is empty
	// means not use config to run, also need to keep cookie
	var config apistructs.AutoTestAPIConfig
	if len(ctx.SDK.Task.Extra.PrivateEnvs[autotest_cookie_keep_after.AutotestApiGlobalConfig]) >= 0 {
		err := json.Unmarshal([]byte(ctx.SDK.Task.Extra.PrivateEnvs[autotest_cookie_keep_after.AutotestApiGlobalConfig]), &config)
		if err != nil {
			return err
		}
	}
	if config.Header == nil {
		config.Header = map[string]string{}
	}
	config.Header[autotest_cookie_keep_after.CookieJar] = cookieJar
	configJson, err := json.Marshal(config)
	if err != nil {
		return err
	}
	ctx.SDK.Task.Extra.PrivateEnvs[autotest_cookie_keep_after.AutotestApiGlobalConfig] = string(configJson)

	err = ctx.SDK.DBClient.UpdatePipelineTaskExtra(ctx.SDK.Task.ID, ctx.SDK.Task.Extra)
	if err != nil {
		return err
	}
	return nil
}

func New() *Plugin {
	var p Plugin
	return &p
}
