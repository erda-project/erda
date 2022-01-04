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
	"fmt"
	"runtime/debug"
	"sync"
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
	AssigneeIDs        []string `json:"assigneeIDs,omitempty"`
	StateIDs           []int64  `json:"states,omitempty"`
	FinishedAtStartEnd []*int64 `json:"finishedAtStartEnd"`
}

func (w *Workbench) GetProjIssueQueries(userID string, projIDs []uint64, limit int) (data map[uint64]IssueUrlQueries, err error) {
	data = make(map[uint64]IssueUrlQueries)
	store := new(sync.Map)
	if limit <= 0 {
		limit = 5
	}
	limitCh := make(chan struct{}, limit)
	wg := sync.WaitGroup{}
	defer close(limitCh)

	for _, v := range projIDs {
		// get
		limitCh <- struct{}{}
		wg.Add(1)

		go func(id uint64) {
			defer func() {
				if err := recover(); err != nil {
					logrus.Errorf("")
					logrus.Errorf("%s", debug.Stack())
				}
				// release
				<-limitCh
				wg.Done()
			}()
			res, err := w.GetIssueQueries(userID, id)
			if err != nil {
				logrus.Errorf("get issue queries failed, id: %v, error: %v", id, err)
				return
			}
			store.Store(id, res)
		}(v)
	}
	// wait all finished
	wg.Wait()

	store.Range(func(k interface{}, v interface{}) bool {
		id, ok := k.(uint64)
		if !ok {
			err = fmt.Errorf("projectID: [uint64], assert failed")
			return false
		}
		queries, ok := v.(IssueUrlQueries)
		if !ok {
			err = fmt.Errorf("IssueUrlQueries, assert failed")
			return false
		}
		data[id] = queries
		return true
	})
	return
}

func (w *Workbench) GetIssueQueries(userID string, projID uint64) (IssueUrlQueries, error) {
	var data IssueUrlQueries
	ids, err := w.GetAllIssueStateIDs(projID)
	if err != nil {
		logrus.Errorf("get issue state id failed, request: %+v, error: %v", projID, err)
		return data, err
	}
	expiredStartEndTime := genExpiredStartEndTime()
	todayExpireStartEndTime := genTodayExpireStartEndTime()
	expiredIssueQuery := Query{
		AssigneeIDs:        []string{userID},
		StateIDs:           ids,
		FinishedAtStartEnd: expiredStartEndTime,
	}
	todayExpireIssueQuery := Query{
		AssigneeIDs:        []string{userID},
		StateIDs:           ids,
		FinishedAtStartEnd: todayExpireStartEndTime,
	}
	undoIssueQuery := Query{
		AssigneeIDs: []string{userID},
		StateIDs:    ids,
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

func (w *Workbench) ListIssueStreams(issueIDs []uint64, limit int) (data map[uint64]apistructs.IssueStream, err error) {
	req := apistructs.IssueStreamPagingRequest{
		PageNo:   1,
		PageSize: 1,
	}
	if limit <= 0 {
		limit = 5
	}

	data = make(map[uint64]apistructs.IssueStream)
	store := new(sync.Map)
	limitCh := make(chan struct{}, limit)
	wg := sync.WaitGroup{}
	defer close(limitCh)

	for _, v := range issueIDs {
		// get
		limitCh <- struct{}{}
		wg.Add(1)
		go func(id uint64) {
			defer func() {
				if err := recover(); err != nil {
					logrus.Errorf("")
					logrus.Errorf("%s", debug.Stack())
				}
				// release
				<-limitCh
				wg.Done()
			}()
			req.IssueID = id
			res, err := w.bdl.GetIssueStreams(req)
			if err != nil {
				logrus.Errorf("get issue streams failed, request: %v, error: %v", req, err)
				return
			}
			if len(res.List) == 0 {
				store.Store(id, "")
			} else {
				store.Store(id, res.List[0])
			}
			logrus.Infof("id: %v, response: %+v", id, res)
		}(v)
	}
	wg.Wait()
	store.Range(func(k interface{}, v interface{}) bool {
		id, ok := k.(uint64)
		if !ok {
			err = fmt.Errorf("issueID: [uint64], assert failed")
			return false
		}
		latestStream, ok := v.(apistructs.IssueStream)
		if !ok {
			err = fmt.Errorf("issueContent, assert failed")
			return false
		}
		data[id] = latestStream
		return true
	})
	return
}
