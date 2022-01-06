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
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/providers/component-protocol/components/list"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
)

func (l *List) GenAppKvInfo(item apistructs.ApplicationDTO) (kvs []list.KvInfo) {
	var isPublic = "privateApp"
	if item.IsPublic {
		isPublic = "publicApp"
	}
	updated := l.UpdatedTime(item.UpdatedAt)
	kvs = []list.KvInfo{
		{
			Icon:  "lock",
			Value: l.sdk.I18n(isPublic),
		},
		{
			Icon:  "xiangmuguanli",
			Tip:   l.sdk.I18n("projectInfo"),
			Value: item.ProjectName,
		},
		{
			Icon:  "list-numbers",
			Tip:   l.sdk.I18n("runtime"),
			Value: strconv.Itoa(int(item.Stats.CountRuntimes)),
			Operations: map[cptype.OperationKey]cptype.Operation{
				list.OpItemClickGoto{}.OpKey(): cputil.NewOpBuilder().
					WithSkipRender(true).
					WithServerDataPtr(list.OpItemClickGotoServerData{
						OpItemBasicServerData: list.OpItemBasicServerData{
							Params: map[string]interface{}{
								"projectId": item.ProjectID,
								"appId":     item.ID,
							},
							Target: "deploy",
						},
					}).
					Build(),
			},
		},
		{
			Icon:  "time",
			Tip:   l.sdk.I18n("updatedAt") + ": " + item.UpdatedAt.String(),
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

func (l *List) UpdatedTime(activeTime time.Time) string {
	var subStr string
	nowTime := time.Now()
	sub := nowTime.Sub(activeTime)
	if int64(sub.Hours()) >= 24*30*12 {
		subStr = strconv.FormatInt(int64(sub.Hours())/(24*30*12), 10) + " " + l.sdk.I18n("yearAgo")
	} else if int64(sub.Hours()) >= 24*30 {
		subStr = strconv.FormatInt(int64(sub.Hours())/(24*30), 10) + " " + l.sdk.I18n("monthAgo")
	} else if int64(sub.Hours()) >= 24 {
		subStr = strconv.FormatInt(int64(sub.Hours())/24, 10) + " " + l.sdk.I18n("dayAgo")
	} else if int64(sub.Hours()) > 0 {
		subStr = strconv.FormatInt(int64(sub.Hours()), 10) + " " + l.sdk.I18n("hourAgo")
	} else if int64(sub.Minutes()) > 0 {
		subStr = strconv.FormatInt(int64(sub.Minutes()), 10) + " " + l.sdk.I18n("minuteAgo")
	} else {
		subStr = l.sdk.I18n("secondAgo")
	}
	return subStr
}
