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

package pipelinesvc

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// SyncAfterSplitTable 表结构拆分后，extra 表里有些字段需要移动
// commit 字段处理：移动到 commit_detail 中，并将 commit 字段置空作为标志位
// org_name 字段处理：移动到 normal_labels 中，简单起见，同样使用 commit 字段为空作为标志位
func (s *PipelineSvc) SyncAfterSplitTable() {

	// 防止启动瞬间瞬时对 db 的压力
	time.Sleep(time.Minute * 5)

	for {
		const retryOnceLimit = 1000
		cols := []string{"pipeline_id", "normal_labels", "commit_detail", "commit", "org_name"}

		var extras []spec.PipelineExtra
		err := s.dbClient.Cols(cols...).
			Where("`commit` != ''").
			Limit(retryOnceLimit).
			Find(&extras)
		if err != nil {
			logrus.Errorf("[alert] failed to sync pipeline_extras after split table, err: %v", err)
			time.Sleep(time.Second * 30)
			continue
		}

		// update
		for i := range extras {
			extra := extras[i]
			// commit
			extra.CommitDetail.CommitID = extra.Commit
			extra.Commit = ""

			// org_name
			if extra.NormalLabels == nil {
				extra.NormalLabels = make(map[string]string)
			}
			orgName := extra.OrgName
			extra.NormalLabels[apistructs.LabelOrgName] = orgName
			extra.OrgName = ""

			_, err := s.dbClient.ID(extra.PipelineID).Cols(cols...).Update(&extra)
			if err != nil {
				logrus.Errorf("[alert] failed to update pipeline_extra when sync after split table, "+
					"pipelineID: %d, commit: %s, org_name: %s, err: %v",
					extra.PipelineID, extra.CommitDetail.CommitID, orgName, err)
				continue
			}
		}

		if len(extras) < retryOnceLimit {
			// stop sync
			return
		}
	}

}
