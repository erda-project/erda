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

package issueGantt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/gantt"
	"github.com/erda-project/erda/pkg/strutil"
)

type Gantt struct {
	CtxBdl protocol.ContextBundle `json:"-"`
	// 从全局状态返回给框架
	Uids []string

	gantt.CommonGantt
}

type State struct {
	// common state
	gantt.State

	IssuePagingRequest apistructs.IssuePagingRequest `json:"issuePagingRequest,omitempty"`
}

func getEdgeTime(issues []apistructs.Issue) (eMin, eMax *time.Time) {
	var edgeMinTime, edgeMaxTime *time.Time
	var minIssue, maxIssue apistructs.Issue
	for _, v := range issues {
		if edgeMinTime == nil && v.PlanStartedAt != nil {
			edgeMinTime = v.PlanStartedAt
			minIssue = v
		}
		if edgeMaxTime == nil && v.PlanFinishedAt != nil {
			edgeMaxTime = v.PlanFinishedAt
			maxIssue = v
		}
		if edgeMinTime != nil && v.PlanStartedAt != nil && v.PlanStartedAt.Before(*edgeMinTime) {
			edgeMinTime = v.PlanStartedAt
			minIssue = v
		}
		if edgeMaxTime != nil && v.PlanFinishedAt != nil && v.PlanFinishedAt.After(*edgeMaxTime) {
			edgeMaxTime = v.PlanFinishedAt
			maxIssue = v
		}
	}
	logrus.Infof("raw edge issues, min:%+v, max:%+v", minIssue, maxIssue)

	now := time.Now()
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	//if edgeMinTime == nil && edgeMaxTime == nil {
	tmp1 := nowDate.AddDate(0, 0, -15)
	edgeMinTime = &tmp1
	tmp2 := nowDate.AddDate(0, 0, 10)
	edgeMaxTime = &tmp2
	//} else if edgeMinTime == nil {
	//	tmp := edgeMaxTime.AddDate(0, 0, -24)
	//	edgeMinTime = &tmp
	//} else if edgeMaxTime == nil {
	//	tmp := edgeMinTime.AddDate(0, 0, 24)
	//	edgeMaxTime = &tmp
	//} else {
	//	if nowDate.After(*edgeMaxTime) {
	//		edgeMaxTime = &now
	//	}
	//	tmp := edgeMinTime.AddDate(0, 0, 24)
	//	if edgeMaxTime.Before(tmp) {
	//		edgeMaxTime = &tmp
	//	}
	//}
	logrus.Infof("edge issues times, min:%+v, max:%+v", *edgeMinTime, *edgeMaxTime)
	return edgeMinTime, edgeMaxTime
}

func (g *Gantt) genProps(edgeMinTime, edgeMaxTime *time.Time) {
	props := gantt.Props{
		Visible: true,
		RowKey:  "id",
		// ClassName: "task-gantt-table",
		Columns: []gantt.PropColumn{
			{Title: "处理人", DataIndex: "user", Width: 160},
			{Title: "标题", DataIndex: "issues",
				TitleTip: []string{
					"事项的甘特图只有确保正确输入截止日期、预计时间才能正常显示",
					"#gray#灰色#>gray#：代表事项截止日期的剩余时间段",
					"#blue#蓝色#>blue#：代表从事项开始时间到当前/事项完成日期的时间段",
					"#red#红色#>red#：代表截止日期到当前/事项完成日期的超时时间段",
				},
			},
		}}
	ganColumn := gantt.PropColumn{Title: "甘特图", DataIndex: "dateRange", TitleRenderType: "gantt", Width: 800}
	data := g.genGanPropColumn(edgeMinTime, edgeMaxTime)
	ganColumn.Data = data
	props.Columns = append(props.Columns, ganColumn)
	g.setProps(props)
}

func (g Gantt) getZhTaskName() string {
	switch g.State.IssueType {
	case apistructs.IssueTypeTask.String():
		return "任务"
	case apistructs.IssueTypeBug.String():
		return "缺陷"
	case apistructs.IssueTypeRequirement.String():
		return "需求"
	default:
		return "任务"
	}
}

// TODO: generate gantt prop column data based on issues
// 长度为25天
func (g *Gantt) genGanPropColumn(edgeMinTime, edgeMaxTime *time.Time) []gantt.PropColumnData {
	start := *edgeMinTime
	end := *edgeMaxTime
	mDays := make(map[uint64][]uint64)
	var months []uint64
	// range: [start, end]
	for !start.Equal(end) {
		m := uint64(start.Month())
		d := uint64(start.Day())
		if months == nil || m != months[len(months)-1] {
			months = append(months, m)
		}
		mDays[m] = append(mDays[m], d)
		start = start.AddDate(0, 0, 1)
	}
	var data []gantt.PropColumnData
	for _, m := range months {
		data = append(data, gantt.PropColumnData{
			Month: m,
			Date:  mDays[m],
		})
	}
	return data
}

func (g *Gantt) setProps(p gantt.Props) {
	g.Props = p
}

func (g *Gantt) genData(issues []apistructs.Issue, edgeMinTime, edgeMaxTime *time.Time) error {
	// group by assignee
	// test
	//now := time.Now()
	//nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	//edgeMinTime = &nowDatedice-openapi-56979f5cb8-vm6pc
	// ===

	aIssues := make(map[string][]apistructs.Issue)
	var uids []string
	for _, v := range issues {
		aIssues[v.Assignee] = append(aIssues[v.Assignee], v)
		uids = append(uids, v.Assignee)
	}
	uInfos := genUserInfo(uids)
	g.setUids(uids)
	// range by assignee
	var data []gantt.DataItem
	autoID := uint64(10000000000000)
	for k, v := range aIssues {
		uInfo := uInfos[k]
		for _, i := range v {
			t := gantt.DataTask{RenderType: "string-list"}
			r := gantt.DateRange{RenderType: "gantt"}
			tInfo := genTaskInfo(i)
			rInfo := genRangeInfo(edgeMinTime, edgeMaxTime, i)
			t.Value = append(t.Value, tInfo)
			r.Value = append(r.Value, rInfo)
			item := gantt.DataItem{ID: autoID + uint64(i.ID), User: *uInfo, Tasks: t, DateRange: r}
			data = append(data, item)
		}
	}
	g.setData(data)
	return nil
}

// TODO: 如果总返回response透出了userInfos, 此处只需要userID即可, 不需要再查一次
func genUserInfo(uids []string) map[string]*gantt.User {
	gUserInfo := make(map[string]*gantt.User)
	for _, v := range uids {
		uid, _ := strconv.ParseUint(v, 10, 64)
		gUserInfo[v] = &gantt.User{
			Value:      uid,
			RenderType: gantt.RenderTypeMemberAvatar,
		}
	}
	//uInfo, err := posthandle.GetUsers(uids, false)
	//if err != nil {
	//	logrus.Warnf("get user info failed, err:%v", err)
	//	return gUserInfo
	//}
	//for k, v := range uInfo {
	//	gUserInfo[k].Name = v.Name
	//	gUserInfo[k].Nick = v.Nick
	//}
	return gUserInfo
}

func genTaskInfo(i apistructs.Issue) gantt.DataTaskValue {
	return gantt.DataTaskValue{
		Text:        i.Title,
		ID:          i.ID,
		Type:        i.Type,
		IterationID: i.IterationID,
	}
}

func fix(x *time.Time, l time.Time, r time.Time) {
	if x.Before(l) {
		*x = l
	} else if x.After(r) {
		*x = r
	}
}

// TODO:
func genRangeInfo(edgeMinTime, edgeMaxTime *time.Time, i apistructs.Issue) gantt.DateRangeValue {
	// 时间（天）
	// offset 偏移时间（天），开始时间相对边界时间

	var trueFinishDate time.Time
	data := gantt.DateRangeValue{Tooltip: i.Title}
	now := time.Now()
	var startTime time.Time
	if i.PlanFinishedAt != nil {
		*i.PlanFinishedAt = (*i.PlanFinishedAt).Add(time.Hour * time.Duration(24))
	}
	// PlanStartedAt为空，向上取整；不为空，向下取整 （单位: 天）
	if i.PlanFinishedAt != nil && i.ManHour.EstimateTime > 0 {
		startTime = (*i.PlanFinishedAt).Add(-time.Minute * time.Duration(i.ManHour.EstimateTime*3))
	}

	if i.ManHour.EstimateTime == 0 || i.PlanFinishedAt == nil {
		return data
	}

	e := edgeMinTime
	s := &startTime
	f := i.PlanFinishedAt
	ts := i.PlanStartedAt
	tf := i.FinishTime
	// 非法数据不显示
	if ts != nil && tf != nil && tf.Before(*ts) {
		return data
	}
	if ts != nil && now.Before(*ts) {
		return data
	}
	if tf != nil && now.Before(*tf) {
		return data
	}

	edgeDate := time.Date(e.Year(), e.Month(), e.Day(), 0, 0, 0, 0, edgeMinTime.Location())
	fedgeData := time.Date(edgeMaxTime.Year(), edgeMaxTime.Month(), edgeMaxTime.Day(), 0, 0, 0, 0, edgeMinTime.Location())
	startDate := time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, edgeMinTime.Location())
	finishDate := time.Date(f.Year(), f.Month(), f.Day(), 0, 0, 0, 0, edgeMinTime.Location())
	nowData := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, edgeMinTime.Location())
	if tf != nil {
		trueFinishDate = time.Date(tf.Year(), tf.Month(), tf.Day(), tf.Hour(), 0, 0, 0, edgeMinTime.Location())
	}

	fix(&startDate, edgeDate, fedgeData)
	fix(&finishDate, edgeDate, fedgeData)
	if tf != nil {
		fix(&trueFinishDate, edgeDate, fedgeData)
	}

	if ts == nil {
		// 无实际开始时间
		if tf == nil {
			// 无实际结束时间
			if now.Before(*s) || now.Equal(*s) {
				// 当前时间 <= 预计开始时间
				data.RestTime = finishDate.Sub(startDate).Hours() / 24.0
				data.Offset = startDate.Sub(edgeDate).Hours() / 24.0
			} else if now.Before(*f) || now.Equal(*f) {
				// 预计开始时间 < 当前时间 <= 预计结束时间
				data.RestTime = finishDate.Sub(nowData).Hours() / 24.0
				data.Offset = nowData.Sub(edgeDate).Hours() / 24.0
			} else {
				// 预计结束时间 < 当前时间
				data.Delay = now.Sub(finishDate).Hours() / 24.0
				data.Offset = finishDate.Sub(edgeDate).Hours() / 24.0
			}
		} else {
			// 有实际结束时间
			if tf.Before(*f) || tf.Equal(*f) {
				// 实际结束时间 <= 预计结束时间
				data.RestTime = finishDate.Sub(trueFinishDate).Hours() / 24.0
				data.Offset = trueFinishDate.Sub(edgeDate).Hours() / 24.0
			} else {
				// 预计结束时间 < 实际结束时间
				data.RestTime = finishDate.Sub(startDate).Hours() / 24.0
				data.Delay = trueFinishDate.Sub(finishDate).Hours() / 24.0
				data.Offset = startDate.Sub(edgeDate).Hours() / 24.0
			}
		}
	} else {
		// 有实际开始时间
		trueStartDate := time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, edgeMinTime.Location())
		fix(&trueStartDate, edgeDate, fedgeData)
		if tf == nil {
			// 无实际结束时间
			if now.Before(*f) || now.Equal(*f) {
				// 实际开始时间 <= 当前时间 <= 预计结束时间
				data.ActualTime = nowData.Sub(trueStartDate).Hours() / 24.0
				data.RestTime = finishDate.Sub(nowData).Hours() / 24.0
				data.Offset = trueStartDate.Sub(edgeDate).Hours() / 24.0
			} else {
				if ts.Before(*f) || ts.Equal(*f) {
					// 实际开始时间 < 预计结束时间 <= 当前时间
					data.ActualTime = finishDate.Sub(trueStartDate).Hours() / 24.0
					data.Delay = nowData.Sub(finishDate).Hours() / 24.0
					data.Offset = trueStartDate.Sub(edgeDate).Hours() / 24.0
				}
			}
		} else {
			// 有实际结束时间
			if tf.Before(*f) || tf.Equal(*f) {
				// 实际开始时间 <= 实际结束时间 <= 计划结束时间
				data.ActualTime = trueFinishDate.Sub(trueStartDate).Hours() / 24.0
				data.RestTime = finishDate.Sub(trueFinishDate).Hours() / 24.0
				data.Offset = trueStartDate.Sub(edgeDate).Hours() / 24.0
			} else {
				if ts.Before(*f) || ts.Equal(*f) {
					// 实际开始时间 < 预计结束时间 <= 实际结束时间
					data.ActualTime = finishDate.Sub(trueStartDate).Hours() / 24.0
					data.Offset = trueStartDate.Sub(edgeDate).Hours() / 24.0
				} else {
					// 预计结束时间 <= 实际开始时间 < 实际结束时间
					data.Offset = finishDate.Sub(edgeDate).Hours() / 24.0
				}
				data.Delay = trueFinishDate.Sub(finishDate).Hours() / 24.0

			}
		}
	}

	ElapsedHour := float64(i.ManHour.ElapsedTime) / (60.0 * 8.0)
	data.Tooltip = fmt.Sprintf("%s (实际用时: %.1f天)", data.Tooltip, ElapsedHour)
	return data
}

func (g *Gantt) setData(data []gantt.DataItem) {
	g.Data = gantt.Data{List: data}
}

func (g *Gantt) setOperations() {
	ops := make(map[apistructs.OperationKey]apistructs.Operation)
	ops[gantt.OpChangePageNo] = apistructs.Operation{Key: gantt.OpChangePageNo.String(), Reload: true}
	g.Operations = ops
}

func (g *Gantt) setDefaultState(operation apistructs.OperationKey) {
	// default state
	defaultState := gantt.State{
		PageNo:   gantt.DefaultPageNo,
		PageSize: gantt.DefaultPageSize,
	}

	switch operation {
	// global default state form url
	case apistructs.InitializeOperation:
		if urlQueryI, ok := g.CtxBdl.InParams[getStateUrlQueryKey()]; ok {
			if urlQueryStr, ok := urlQueryI.(string); ok && urlQueryStr != "" {
				var urlState gantt.State
				b, err := base64.StdEncoding.DecodeString(urlQueryStr)
				if err != nil {
					logrus.Warnf("decode url query failed, request:%v, err:%v", urlQueryStr, err)
				}
				if err := json.Unmarshal(b, &urlState); err != nil {
					logrus.Warnf("get url state failed, request:%v, err:%v", urlState, err)
				}
				if err == nil {
					defaultState = urlState
				}
			}
		}
	}

	g.State.PageSize = defaultState.PageSize
	g.State.PageNo = defaultState.PageNo
}

func (g *Gantt) setStateTotal(total uint64) {
	g.State.Total = total
}

func (g *Gantt) setUids(uids []string) {
	uids = strutil.DedupSlice(uids, true)
	g.Uids = uids
}
