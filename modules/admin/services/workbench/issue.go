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

package workbench

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

type IssueUrlQueries struct {
	ExpiredQuery     string
	TodayExpireQuery string
	UndoQuery        string
}

type Query struct {
	StateIDs           []int64  `json:"states,omitempty"`
	FinishedAtStartEnd []*int64 `json:"finishedAtStartEnd"`
}

func (w *Workbench) GetIssueQueries(projID uint64) (IssueUrlQueries, error) {
	var data IssueUrlQueries
	ids, err := w.GetAllIssueStateIDs(projID)
	if err != nil {
		logrus.Errorf("get issue state id failed, request: %+v, error: %v", projID, err)
		return data, err
	}
	expiredStartEndTime := genExpiredStartEndTime()
	todayExpireStartEndTime := genTodayExpireStartEndTime()
	expiredIssueQuery := Query{
		StateIDs:           ids,
		FinishedAtStartEnd: expiredStartEndTime,
	}
	todayExpireIssueQuery := Query{
		StateIDs:           ids,
		FinishedAtStartEnd: todayExpireStartEndTime,
	}
	undoIssueQuery := Query{
		StateIDs: ids,
	}
	expiredIssueQueryStr, _ := encodeQuery(expiredIssueQuery)
	todayExpireIssueQueryStr, _ := encodeQuery(todayExpireIssueQuery)
	undoIssueQueryStr, _ := encodeQuery(undoIssueQuery)
	data = IssueUrlQueries{
		ExpiredQuery:     expiredIssueQueryStr,
		TodayExpireQuery: todayExpireIssueQueryStr,
		UndoQuery:        undoIssueQueryStr,
	}
	return data, nil
}

func encodeQuery(q Query) (string, error) {
	b, err := json.Marshal(q)
	if err != nil {
		logrus.Errorf("marshal query failed, request: %+v, error:%v", q, err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (w *Workbench) GetAllIssueStateIDs(projID uint64) ([]int64, error) {
	var stateIDs []int64

	statesBlMap := map[apistructs.IssueStateBelong]bool{
		apistructs.IssueStateBelongOpen:     true,
		apistructs.IssueStateBelongWorking:  true,
		apistructs.IssueStateBelongWontfix:  true,
		apistructs.IssueStateBelongReopen:   true,
		apistructs.IssueStateBelongResolved: true,
	}
	types := []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug}

	for _, v := range types {
		req := apistructs.IssueStateRelationGetRequest{
			ProjectID: projID,
			IssueType: v,
		}
		r, err := w.bdl.GetIssueStateBelong(req)
		if err != nil {
			logrus.Errorf("get issue state failed, request: %+v, error: %v", req, err)
			return nil, err
		}
		for _, v := range r {
			// if in target stateBelong, get its ids
			if statesBlMap[v.StateBelong] {
				for i := range v.States {
					stateIDs = append(stateIDs, v.States[i].ID)
				}
			}
		}

	}
	return strutil.DedupInt64Slice(stateIDs), nil
}

func genExpiredStartEndTime() []*int64 {
	end := yesterdayEndTime()
	return []*int64{nil, end}
}

func genTodayExpireStartEndTime() []*int64 {
	start := todayStartTime()
	end := todayEndTime()
	return []*int64{start, end}
}

func todayStartTime() *int64 {
	now := time.Now()
	tm := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	msTime := tm.UnixNano() / 1e6
	return &msTime
}

func todayEndTime() *int64 {
	now := time.Now()
	tm := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	msTime := tm.UnixNano() / 1e6
	return &msTime
}

func yesterdayEndTime() *int64 {
	now := time.Now()
	yes := now.AddDate(0, 0, -1)
	tm := time.Date(yes.Year(), yes.Month(), yes.Day(), 23, 59, 59, 0, now.Location())
	msTime := tm.UnixNano() / 1e6
	return &msTime
}
