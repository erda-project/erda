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

// Package issue 封装 事件 相关操作
package issue

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/strutil"
)

// Issue 事件操作封装
type Issue struct {
	db *dao.DBClient
}

// Option 定义 Issue 配置选项
type Option func(*Issue)

// New 新建 Issue 实例
func New(options ...Option) *Issue {
	itr := &Issue{}
	for _, op := range options {
		op(itr)
	}
	return itr
}

func WithIssueDBClient(db *dao.DBClient) Option {
	return func(issue *Issue) {
		issue.db = db
	}
}

func (svc *Issue) GetIssuesByStates(req apistructs.WorkbenchRequest) (map[uint64]*apistructs.WorkbenchProjectItem, error) {
	stats, err := svc.db.GetIssueExpiryStatusByProjects(req)
	if err != nil {
		return nil, err
	}
	projectMap := make(map[uint64]*apistructs.WorkbenchProjectItem)
	for _, i := range stats {
		if _, ok := projectMap[i.ProjectID]; !ok {
			projectMap[i.ProjectID] = &apistructs.WorkbenchProjectItem{}
		}
		item := projectMap[i.ProjectID]
		num := int(i.IssueNum)
		switch i.ExpiryStatus {
		case dao.ExpireTypeUndefined:
			item.UnSpecialIssueNum = num
		case dao.ExpireTypeExpired:
			item.ExpiredIssueNum = num
		case dao.ExpireTypeExpireIn1Day:
			item.ExpiredOneDayNum = num
		case dao.ExpireTypeExpireIn2Days:
			item.ExpiredTomorrowNum = num
		case dao.ExpireTypeExpireIn7Days:
			item.ExpiredSevenDayNum = num
		case dao.ExpireTypeExpireIn30Days:
			item.ExpiredThirtyDayNum = num
		case dao.ExpireTypeExpireInFuture:
			item.FeatureDayNum = num
		}
		item.TotalIssueNum += num
	}
	return projectMap, nil
}

// getRelatedIDs 当同时通过lable和issue关联关系过滤时，需要取交集
func getRelatedIDs(lableRelationIDs []int64, issueRelationIDs []int64, isLabel, isIssue bool) []int64 {
	// 取交集
	if isLabel && isIssue {
		return strutil.IntersectionInt64Slice(lableRelationIDs, issueRelationIDs)
	}
	if isLabel {
		return lableRelationIDs
	}
	if isIssue {
		return issueRelationIDs
	}
	return nil
}
