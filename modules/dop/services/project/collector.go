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

package project

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

func (p *Project) CollectApplicationsResource() (bool, error) {
	l := logrus.WithField("func", "*Project.CollectApplicationsResource")

	// 1) 查出所有的 project
	projects, err := p.bdl.GetAllProjects()
	if err != nil {
		l.WithError(err).Errorln("failed to GetAllProjects")
		return false, err
	}

	// 2) 对每一个 project 查询其下 apps 的资源
	for _, project := range projects {
		resources, apiError := p.ApplicationsResources(context.Background(), &apistructs.ApplicationsResourcesRequest{
			OrgID:     strconv.FormatUint(project.OrgID, 32),
			UserID:    "0",
			ProjectID: strconv.FormatUint(project.ID, 3),
			Query:     &apistructs.ApplicationsResourceQuery{PageNo: "1", PageSize: "100"},
		})
		if apiError != nil {
			l.WithError(err).WithField("projectID", project.ID).Errorln("failed to ApplicationsResources")
			continue
		}
		if resources == nil || len(resources.List) == 0 {
			l.WithField("projectID", project.ID).Warnln("no application items in the project ")
			continue
		}
		for _, app := range resources.List {
			var daily = apistructs.ApplicationResourceDailyModel{
				ProjectID:              project.ID,
				ApplicationID:          app.ID,
				ApplicationName:        app.Name,
				ApplicationDisplayName: app.DisplayName,
				ProdCPURequest:         app.ProdCPURequest,
				ProdMemRequest:         app.ProdMemRequest,
				ProdPodsCount:          app.ProdPodsCount,
				StagingCPURequest:      app.StagingCPURequest,
				StagingMemRequest:      app.StagingMemRequest,
				StagingPodsCount:       app.StagingPodsCount,
				TestCPURequest:         app.TestCPURequest,
				TestMemRequest:         app.TestMemRequest,
				TestPodsCount:          app.TestPodsCount,
				DevCPURequest:          app.DevCPURequest,
				DevMemRequest:          app.DevMemRequest,
				DevPodsCount:           app.DevPodsCount,
			}
			var existsRecord apistructs.ApplicationResourceDailyModel
			err := p.db.Where("created_at >= ? AND created_at < ?",
				time.Now().Format("2006-01-02 00:00:00"),
				time.Now().Add(time.Hour*24).Format("2006-01-02 00:00:00")).
				First(&existsRecord, map[string]interface{}{"application_id": app.ID}).
				Error
			switch {
			case err == nil:
				daily.ID = existsRecord.ID
				err = p.db.Debug().Save(&daily).Error
			case gorm.IsRecordNotFoundError(err):
				err = p.db.Create(&daily).Error
			default:
				err = errors.Wrap(err, "failed to First existsRecord")
			}
			if err != nil {
				dailyContent, _ := json.Marshal(daily)
				l.WithError(err).WithField("record", string(dailyContent)).Warnln()
			}
		}
	}
	return false, nil
}
