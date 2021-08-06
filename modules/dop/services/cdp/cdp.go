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

// Package cdp pipeline相关的结构信息
package cdp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/conf"
)

// CDP pipeline 结构体
type CDP struct {
	bdl *bundle.Bundle
}

// Option CDP 配置选项
type Option func(*CDP)

// New CDP service
func New(options ...Option) *CDP {
	r := &CDP{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(f *CDP) {
		f.bdl = bdl
	}
}

func (cdp *CDP) CdpNotifyProcess(pipelineEvent *apistructs.PipelineInstanceEvent) error {
	pipelineData := pipelineEvent.Content
	orgID, err := strconv.ParseInt(pipelineEvent.OrgID, 10, 64)
	if err != nil {
		return err
	}
	pipelineDetail, err := cdp.bdl.GetPipeline(pipelineData.PipelineID)
	if err != nil {
		return err
	}
	if strings.Index(pipelineData.Source, "cdp-") == 0 {
		logrus.Infof("cdp event: %+v", pipelineEvent)
		workflowID, ok := pipelineDetail.Labels["CDP_WF_ID"]
		if !ok {
			return fmt.Errorf("failed to get workflow id pipelineID:%d", pipelineData.PipelineID)
		}
		workflowName, ok := pipelineDetail.Labels["CDP_WF_NAME"]
		if !ok {
			return fmt.Errorf("failed to get workflow name pipelineID:%d", pipelineData.PipelineID)
		}
		fdpNotifyLabel, fdpOK := pipelineDetail.Labels["FDP_NOTIFY_LABLE"]
		if !ok {
			logrus.Infof("failed to get FDP_NOTIFY_LABLE: %d", pipelineData.PipelineID)
		}
		logrus.Debugf("fdpNotifyLabel and pipelineID is: %s, %d", fdpNotifyLabel, pipelineData.PipelineID)
		notifyDetails, err := cdp.bdl.QueryNotifiesBySource(pipelineEvent.OrgID, "workflow", workflowID,
			pipelineData.Status, fdpNotifyLabel, pipelineDetail.ClusterName)
		if err != nil {
			return err
		}
		// 一条事件多个接收者时，一个接收者发生错误后，需要继续进行下去
		for _, notifyDetail := range notifyDetails {
			if notifyDetail.NotifyGroup == nil {
				continue
			}
			var notifyItem *apistructs.NotifyItem
			for _, item := range notifyDetail.NotifyItems {
				if item.Name == pipelineData.Status {
					notifyItem = item
				}
			}
			if notifyItem == nil {
				continue
			}

			params := map[string]string{
				"pipelineID":     strconv.FormatUint(pipelineData.PipelineID, 10),
				"workflowID":     workflowID,
				"workflowName":   workflowName,
				"notifyItemName": notifyItem.DisplayName,
				"clusterName":    pipelineData.ClusterName,
			}
			if fdpOK {
				// fdp 的通知历史记录label
				params["fdpNotifyLabel"] = fdpNotifyLabel
			}

			//失败情况尝输出错误日志
			if notifyItem.Name == "Failed" {
				time.Sleep(time.Second * 5)
				failedDetailLogs, err := cdp.getFailedTaskLogs(pipelineDetail)
				if err != nil {
					logrus.Errorf("get cdp workflow's failed log err: %v", err)
					continue
				}
				params["failedDetail"] = failedDetailLogs

			}
			// 不使用notifyItemID是为了支持第三方通知项,如监控
			eventboxReqContent := apistructs.GroupNotifyContent{
				SourceName:            workflowName,
				SourceType:            "workflow",
				SourceID:              workflowID,
				NotifyName:            notifyDetail.Name,
				NotifyItemDisplayName: notifyItem.DisplayName,
				Channels:              []apistructs.GroupNotifyChannel{},
				Label:                 notifyItem.Label,
				CalledShowNumber:      notifyItem.CalledShowNumber,
				ClusterName:           pipelineDetail.ClusterName,
				OrgID:                 orgID,
			}

			err := cdp.bdl.CreateGroupNotifyEvent(apistructs.EventBoxGroupNotifyRequest{
				Sender:        "adapter",
				GroupID:       notifyDetail.NotifyGroup.ID,
				Channels:      notifyDetail.Channels,
				NotifyItem:    notifyItem,
				NotifyContent: &eventboxReqContent,
				Params:        params,
			})
			if err != nil {
				logrus.Errorf("create group notify event failed err: %v", err)
				continue
			}
		}
	} else {
		logrus.Infof("pipeline event: %+v", pipelineEvent)
		//普通pipeline
		eventName := "pipeline_" + strings.ToLower(pipelineData.Status)
		sourceType := "app"
		sourceID := strconv.FormatUint(pipelineDetail.ApplicationID, 10)
		notifyDetails, err := cdp.bdl.QueryNotifiesBySource(pipelineEvent.OrgID,
			sourceType, sourceID, eventName, "")
		if err != nil {
			return err
		}
		// 一条事件多个接收者时，一个接收者发生错误后，需要继续进行下去
		for _, notifyDetail := range notifyDetails {
			if notifyDetail.NotifyGroup == nil {
				continue
			}
			notifyItem := notifyDetail.NotifyItems[0]
			params := map[string]string{
				"pipelineID":     strconv.FormatUint(pipelineData.PipelineID, 10),
				"notifyItemName": notifyItem.DisplayName,
				"appID":          strconv.FormatUint(pipelineDetail.ApplicationID, 10),
				"appName":        pipelineDetail.ApplicationName,
				"projectID":      strconv.FormatUint(pipelineDetail.ProjectID, 10),
				"projectName":    pipelineDetail.ProjectName,
				"orgName":        pipelineDetail.OrgName,
				"branch":         pipelineDetail.Branch,
				"uiPublicURL":    conf.UIPublicURL(),
			}
			//失败情况尝输出错误日志
			if notifyItem.Name == "pipeline_failed" {
				// 等待日志采集
				time.Sleep(time.Second * 5)
				failedDetailLogs, err := cdp.getFailedTaskLogs(pipelineDetail)
				if err != nil {
					logrus.Errorf("get cdp workflow's failed log err: %v", err)
					failedDetailLogs = "Log cannot be displayed"
				}
				params["failedDetail"] = failedDetailLogs
			}
			// 不使用notifyItemID是为了支持第三方通知项,如监控
			eventboxReqContent := apistructs.GroupNotifyContent{
				SourceName:            strconv.FormatUint(pipelineData.PipelineID, 10),
				SourceType:            sourceType,
				SourceID:              sourceID,
				NotifyName:            notifyDetail.Name,
				NotifyItemDisplayName: notifyItem.DisplayName,
				Channels:              []apistructs.GroupNotifyChannel{},
				ClusterName:           pipelineDetail.ClusterName,
				CalledShowNumber:      notifyItem.CalledShowNumber,
				Label:                 notifyItem.Label,
				OrgID:                 orgID,
			}

			err := cdp.bdl.CreateGroupNotifyEvent(apistructs.EventBoxGroupNotifyRequest{
				Sender:        "adapter",
				GroupID:       notifyDetail.NotifyGroup.ID,
				Channels:      notifyDetail.Channels,
				NotifyItem:    notifyItem,
				NotifyContent: &eventboxReqContent,
				Params:        params,
			})
			if err != nil {
				logrus.Errorf("get cdp workflow's failed log err: %v", err)
				continue
			}
		}
	}
	return nil
}

func (cdp *CDP) getFailedTaskLogs(pipelineDetail *apistructs.PipelineDetailDTO) (string, error) {
	failedTasks := map[string]string{}
	for _, stage := range pipelineDetail.PipelineStages {
		for _, task := range stage.PipelineTasks {
			if task.Status == "Failed" {
				failedTasks[task.Name] = task.Extra.UUID
			}
		}
	}
	failedDetailLogs := ""
	for taskName, taskID := range failedTasks {
		stdError, err := cdp.bdl.GetLog(apistructs.DashboardSpotLogRequest{
			ID:     taskID,
			Source: apistructs.DashboardSpotLogSourceJob,
			Stream: apistructs.DashboardSpotLogStreamStderr,
			Count:  -50,
			Start:  0,
			End:    time.Duration(time.Now().UnixNano()),
		})
		if err != nil {
			return "", err
		}
		stdOut, err := cdp.bdl.GetLog(apistructs.DashboardSpotLogRequest{
			ID:     taskID,
			Source: apistructs.DashboardSpotLogSourceJob,
			Stream: apistructs.DashboardSpotLogStreamStdout,
			Count:  -50,
			Start:  0,
			End:    time.Duration(time.Now().UnixNano()),
		})
		if err != nil {
			return "", err
		}
		stderrLines := ""
		stdoutLines := ""
		for _, line := range stdError.Lines {
			stderrLines += line.Content + "\n"
		}
		for _, line := range stdOut.Lines {
			stdoutLines += line.Content + "\n"
		}

		if len(stdoutLines) > 0 {
			failedDetailLogs += fmt.Sprintf(`<p>%s stdout:</p>
<pre style="background:#E8E8E8"><code>%s</pre></code>`, taskName, stdoutLines)
		}

		if len(stderrLines) > 0 {
			failedDetailLogs += fmt.Sprintf(`<p>%s stderr:</p>
<pre style="background:#E8E8E8"><code>%s</pre></code>`, taskName, stderrLines)
		}
	}
	return failedDetailLogs, nil
}
