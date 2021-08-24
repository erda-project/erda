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

package executeInfo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentFileInfo struct {
	CtxBdl protocol.ContextBundle

	CommonFileInfo
}

type CommonFileInfo struct {
	Version    string                                           `json:"version,omitempty"`
	Name       string                                           `json:"name,omitempty"`
	Type       string                                           `json:"type,omitempty"`
	Props      map[string]interface{}                           `json:"props,omitempty"`
	State      State                                            `json:"state,omitempty"`
	Operations map[apistructs.OperationKey]apistructs.Operation `json:"operations,omitempty"`
	Data       map[string]interface{}                           `json:"data,omitempty"`
}

type PropColumn struct {
	Label    string `json:"label"`
	ValueKey string `json:"valueKey"`
}

type State struct {
	PipelineID uint64 `json:"pipelineId"`
}

type reportNew struct {
	APIFailedNum  int64 `json:"apiFailedNum"`
	APINotExecNum int64 `json:"apiNotExecNum"`
	APISuccessNum int64 `json:"apiSuccessNum"`
	APITotalNum   int64 `json:"apiTotalNum"`
}

func (a *ComponentFileInfo) Import(c *apistructs.Component) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, a); err != nil {
		return err
	}
	return nil
}

// GenComponentState 获取state
func (i *ComponentFileInfo) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	fmt.Println(state)
	i.State = state
	return nil
}

func (i *ComponentFileInfo) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) {
	if c.Data == nil {
		d := make(apistructs.ComponentData)
		c.Data = d
	}
	(*c).Data["data"] = i.Data
	c.Props = i.Props

}

func (i *ComponentFileInfo) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := i.Import(c); err != nil {
		logrus.Errorf("import component failed, err:%v", err)
		return err
	}

	i.CtxBdl = ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	defer func() {
		fail := i.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()
	if i.State.PipelineID > 0 {
		rsp, err := i.CtxBdl.Bdl.GetPipeline(i.State.PipelineID)
		if err != nil {
			return err
		}
		if rsp.TimeBegin != nil && (rsp.TimeEnd != nil || rsp.TimeUpdated != nil) && rsp.Status.IsEndStatus() {
			var timeLayoutStr = "2006-01-02 15:04:05" //go中的时间格式化必须是这个时间
			if rsp.TimeEnd == nil {
				rsp.TimeEnd = rsp.TimeUpdated
			}
			t := rsp.TimeEnd.Sub(*rsp.TimeBegin)
			h := strconv.FormatInt(int64(t.Hours()), 10) + ":"
			if t.Hours() < 10 {
				h = "0" + strconv.FormatInt(int64(t.Hours()), 10) + ":"
			}
			i.Data = map[string]interface{}{
				"pipelineID": i.State.PipelineID,
				"status":     rsp.Status.ToDesc(),
				"time":       h + time.Unix(int64(t.Seconds())-8*3600, 0).Format("04:05"),
				"timeBegin":  rsp.TimeBegin.Format(timeLayoutStr),
				"timeEnd":    rsp.TimeEnd.Format(timeLayoutStr),
			}
		} else if rsp.TimeBegin != nil {
			var timeLayoutStr = "2006-01-02 15:04:05" //go中的时间格式化必须是这个时间
			i.Data = map[string]interface{}{
				"pipelineID": i.State.PipelineID,
				"status":     rsp.Status.ToDesc(),
				"timeBegin":  rsp.TimeBegin.Format(timeLayoutStr),
			}
		} else {
			i.Data = map[string]interface{}{
				"pipelineID": i.State.PipelineID,
				"status":     rsp.Status.ToDesc(),
			}
		}
		if rsp.Status == apistructs.PipelineStatusStopByUser {
			i.Data["status"] = "用户取消"
		}
		if rsp.Status == apistructs.PipelineStatusNoNeedBySystem {
			i.Data["status"] = "无需执行"
		}
		res, err := i.CtxBdl.Bdl.GetPipelineReportSet(i.State.PipelineID, []string{"api-test"})
		if err != nil {
			return err
		}
		fmt.Println(res)
		if res != nil && len(res.Reports) > 0 && res.Reports[0].Meta != nil {
			value, err := json.Marshal(res.Reports[0].Meta)
			if err != nil {
				i.Data["autoTestExecPercent"] = "-"
				i.Data["autoTestSuccessPercent"] = "-"
				goto Label
			}
			var report reportNew
			err = json.Unmarshal(value, &report)
			if err != nil {
				i.Data["autoTestExecPercent"] = "-"
				i.Data["autoTestSuccessPercent"] = "-"
				goto Label
			}
			i.Data["autoTestNum"] = report.APITotalNum
			if report.APITotalNum == 0 {
				i.Data["autoTestExecPercent"] = "0.00%"
				i.Data["autoTestSuccessPercent"] = "0.00%"
			} else {
				i.Data["autoTestExecPercent"] = strconv.FormatFloat(100-float64(report.APINotExecNum)/float64(report.APITotalNum)*100, 'f', 2, 64) + "%"
				i.Data["autoTestSuccessPercent"] = strconv.FormatFloat(float64(report.APISuccessNum)/float64(report.APITotalNum)*100, 'f', 2, 64) + "%"
			}
		}
	}
Label:
	i.Props = make(map[string]interface{})
	i.Props["fields"] = []PropColumn{
		{
			Label:    "流水线ID",
			ValueKey: "pipelineID",
		},
		{
			Label:    "状态",
			ValueKey: "status",
		},
		{
			Label:    "时长",
			ValueKey: "time",
		},
		{
			Label:    "开始时间",
			ValueKey: "timeBegin",
		},
		{
			Label:    "结束时间",
			ValueKey: "timeEnd",
		},
		{
			Label:    "接口总数",
			ValueKey: "autoTestNum",
		},
		{
			Label:    "接口执行率",
			ValueKey: "autoTestExecPercent",
		},
		{
			Label:    "接口通过率",
			ValueKey: "autoTestSuccessPercent",
		},
	}

	i.RenderProtocol(c, gs)
	return
}

func (a *ComponentFileInfo) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(a.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	propValue, err := json.Marshal(a.Props)
	if err != nil {
		return err
	}
	var props interface{}
	err = json.Unmarshal(propValue, &props)
	if err != nil {
		return err
	}

	c.Props = props
	c.State = state
	c.Type = a.Type
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentFileInfo{}
}
