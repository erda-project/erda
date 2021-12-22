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
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

func (w *Workbench) GetAppNum(identity apistructs.Identity, query string) (int, error) {
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ApplicationListRequest{
		PageNo:   1,
		PageSize: 1,
		Query:    query,
		IsSimple: true,
	}
	appDTO, err := w.bdl.GetAllMyApps(identity.UserID, uint64(orgID), req)
	if err != nil {
		return 0, err
	}
	if appDTO == nil {
		return 0, nil
	}
	return appDTO.Total, nil
}

// ListAppWbData
// default set pageSize/pageNo; when need query, set query field
func (w *Workbench) ListAppWbData(identity apistructs.Identity, req apistructs.ApplicationListRequest, limit int) (*apistructs.AppWorkbenchResponseData, error) {
	var (
		appIDs []uint64
		data   apistructs.AppWorkbenchResponseData
	)
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return nil, err
	}
	req.OrderBy = "name"
	req.IsSimple = true

	// list app
	appRes, err := w.bdl.GetAllMyApps(identity.UserID, uint64(orgID), req)
	if err != nil {
		return nil, err
	}

	for i := range appRes.List {
		appIDs = append(appIDs, appRes.List[i].ID)
	}

	if len(appIDs) == 0 {
		return nil, nil
	}

	// list app related runtime
	runtimeRes, err := w.bdl.ListRuntimesGroupByApps(uint64(orgID), identity.UserID, appIDs)
	if err != nil {
		logrus.Errorf("list runtime group by apps failed, appIDs: %v, error: %v", appIDs, err)
		return nil, err
	}

	// list app related open mr
	mrResult, err := w.ListOpenMrWithLimitRate(identity, appIDs, limit)
	if err != nil {
		logrus.Errorf("list open mr failed, appIDs: %v, error: %v", appIDs, err)
		return nil, err
	}

	// construct AppWorkBenchItem
	data.TotalApps = appRes.Total
	for i := range appRes.List {
		data.List = append(data.List, apistructs.AppWorkBenchItem{
			ApplicationDTO: appRes.List[i],
			AppRuntimeNum:  len(runtimeRes[appRes.List[i].ID]),
			AppOpenMrNum:   mrResult[appRes.List[i].ID],
		})
	}
	return &data, nil
}

func (w *Workbench) ListSubAppWbData(identity apistructs.Identity, limit int) (*apistructs.AppWorkbenchResponseData, error) {
	var (
		idList []uint64
	)
	subList, err := w.bdl.ListSubscribes(identity.UserID, identity.OrgID, apistructs.GetSubscribeReq{Type: apistructs.AppSubscribe})
	if err != nil {
		return nil, err
	}
	for _, v := range subList.List {
		idList = append(idList, v.TypeID)
	}

	req := apistructs.ApplicationListRequest{
		PageNo:        1,
		PageSize:      len(idList),
		ApplicationID: idList,
	}

	return w.ListAppWbData(identity, req, limit)
}

// ListOpenMrWithLimitRate
// TODO: parallel query gittar may have performance issue, need to switch to close it wen necessary
func (w *Workbench) ListOpenMrWithLimitRate(identity apistructs.Identity, appIDs []uint64, limit int) (result map[uint64]int, err error) {
	req := apistructs.GittarQueryMrRequest{
		State: "open",
		Page:  1,
		Size:  0,
	}
	if limit <= 0 {
		limit = 5
	}

	result = make(map[uint64]int)
	store := new(sync.Map)
	limitCh := make(chan struct{}, limit)
	defer close(limitCh)

	for _, v := range appIDs {
		// get
		limitCh <- struct{}{}
		go func(appID uint64) {
			defer func() {
				if err := recover(); err != nil {
					logrus.Errorf("")
					logrus.Errorf("%s", debug.Stack())
				}
				// release
				<-limitCh
			}()
			res, err := w.bdl.ListMergeRequest(appID, identity.UserID, req)
			if err != nil {
				logrus.Warnf("list merget request failed, appID: %v, error: %v", appID, err)
			}
			if res == nil {
				store.Store(appID, 0)
			} else {
				store.Store(appID, res.Total)
			}
		}(v)
	}
	store.Range(func(k interface{}, v interface{}) bool {
		appID, ok := k.(uint64)
		if !ok {
			err = fmt.Errorf("appID type: [int64], assert failed")
			return false
		}
		openMrNum, ok := v.(int)
		if !ok {
			err = fmt.Errorf("openMrNum type: [int], assert failed")
			return false
		}
		result[appID] = openMrNum
		return true
	})
	return
}
