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

package list

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/component-protocol/components/list"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/common"
)

func (l *List) GenAppKvInfo(item apistructs.ApplicationDTO, mrCount int) (kvs []list.KvInfo) {
	var isPublic = "privateApp"
	var publicIcon = "private"
	if item.IsPublic {
		isPublic = "publicApp"
		publicIcon = "public"
	}
	updated := common.UpdatedTime(l.sdk.Ctx, item.UpdatedAt)
	kvs = []list.KvInfo{
		{
			Icon:  publicIcon,
			Value: l.sdk.I18n(isPublic),
			Tip:   l.sdk.I18n("publicProperty"),
		},
		{
			Icon:  "list-numbers",
			Tip:   l.sdk.I18n("openMrCount"),
			Value: strconv.Itoa(mrCount),
			Operations: map[cptype.OperationKey]cptype.Operation{
				list.OpItemClickGoto{}.OpKey(): cputil.NewOpBuilder().
					WithSkipRender(true).
					WithServerDataPtr(list.OpItemClickGotoServerData{
						OpItemBasicServerData: list.OpItemBasicServerData{
							Params: map[string]interface{}{
								"projectId": item.ProjectID,
								"appId":     item.ID,
							},
							Target: "appOpenMr",
						},
					}).
					Build(),
			},
		},
		{
			Icon:  "time",
			Tip:   l.sdk.I18n("updatedAt") + ": " + item.UpdatedAt.Format("2006-01-02 15:04:05"),
			Value: updated,
		},
		{
			Icon:  "category-management",
			Tip:   l.sdk.I18n("appType"),
			Value: l.sdk.I18n("appMode" + item.Mode),
		},
	}
	return
}

func (l *List) ListOpenMrWithLimitRate(identity apistructs.Identity, appIDs []uint64, limit int) (result map[uint64]int, err error) {
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
	wg := sync.WaitGroup{}
	defer close(limitCh)

	for _, v := range appIDs {
		limitCh <- struct{}{}
		wg.Add(1)
		go func(appID uint64) {
			defer func() {
				if err := recover(); err != nil {
					logrus.Errorf("")
					logrus.Errorf("%s", debug.Stack())
				}
				<-limitCh
				wg.Done()
			}()
			res, err := l.bdl.ListMergeRequest(appID, identity.UserID, req)
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
	wg.Wait()
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
