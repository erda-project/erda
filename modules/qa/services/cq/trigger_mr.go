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

package cq

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

func (cq *CQ) TriggerByMR(mr apistructs.MergeRequestInfo) (uint64, error) {
	//// 目标分支是否是保护分支
	//if mr.TargetBranchRule == nil {
	//	logrus.Warnf("MR CQ: no need analyze, reason: target branch doesn't have branch rule, mr info: %s", printMRInfo(mr))
	//	return 0, nil
	//}
	//if !mr.TargetBranchRule.IsProtect {
	//	logrus.Debugf("MR CQ: no need analyze, reason: target branch is not protected, mr info: %s", printMRInfo(mr))
	//}

	// TODO
	// 查询目标分支开关: 流水线配置
	// 查询分支规则，找到目标分支对应的环境
	// 通过环境找到流水线配置是否存在指定的 KEY: MR_CQ=true
	enable := true

	if !enable {
		logrus.Debugf("CQ MR: no need analyze, reason: MR CQ not enabled, mr info: %s", printMRInfo(mr))
	}

	// 构造流水线执行
	return cq.Analyze(CQRequest{
		AppID:    uint64(mr.AppID),
		Commit:   mr.SourceSha,
		Language: LanguageGo, // 当前只支持 dice go
	})
}

func printMRInfo(mr apistructs.MergeRequestInfo) string {
	return fmt.Sprintf("appID: %d, mr title: %s, source branch: %s(%s), target branch: %s(%s), author: %s, assignee: %s",
		mr.AppID, mr.Title, mr.SourceBranch, mr.SourceSha, mr.TargetBranch, mr.TargetSha, mr.AuthorId, mr.AssigneeId)
}
