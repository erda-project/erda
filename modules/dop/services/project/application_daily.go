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

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/apistructs/echarts"
	calcu "github.com/erda-project/erda/pkg/resourcecalculator"
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

func (p *Project) GetApplicationTrend(ctx context.Context, request *apistructs.GetResourceApplicationTrendReq) (*echarts.Histogram, error) {
	var (
		l                = logrus.WithField("func", "*Project.GetApplicationTrend")
		langCodes        = ctx.Value("Lang").(i18n.LanguageCodes)
		td               = new(echarts.Histogram)
		applicationID, _ = request.Query.GetApplicationID()
		start, _         = request.Query.GetStart()
		end, _           = request.Query.GetEnd()
		dailies          []*apistructs.ApplicationResourceDailyModel
	)
	td.Name = p.trans.Text(langCodes, "applications trend")
	td.XAxis = echarts.XAxis{
		Type: "category",
		Data: nil,
	}
	td.YAxis = echarts.YAxis{
		Type: "value",
		Name: "",
	}
	td.Series = []echarts.HistogramSerie{{
		Data: nil,
		Name: p.trans.Text(langCodes, "prod"),
		Type: "bar",
	}, {
		Data: nil,
		Name: p.trans.Text(langCodes, "staging"),
		Type: "bar",
	}, {
		Data: nil,
		Name: p.trans.Text(langCodes, "test"),
		Type: "bar",
	}, {
		Data: nil,
		Name: p.trans.Text(langCodes, "dev"),
		Type: "bar",
	}}

	startTime := time.Unix(int64(start)/1e3, int64(start)%1e3*1e6)
	endTime := time.Unix(int64(end)/1e3, int64(end)%1e3*1e6)
	err := p.db.Where("updated_at >= ? and created_at <= ?",
		startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05")).
		Where("application_id = ?", applicationID).
		Find(&dailies).Error
	if gorm.IsRecordNotFoundError(err) {
		return td, nil
	}
	if err != nil {
		l.WithError(err).Errorln("failed to Find dailies")
		return nil, err
	}
	var records = make(map[string]apistructs.ApplicationResourceDailyModel)
	for _, record := range dailies {
		records[record.CreatedAt.Format("2006-01-02")] = *record
	}
	for day := startTime; day.Before(endTime) && day.Before(time.Now()); day = day.Add(time.Hour * 24) {
		today := day.Format("2006-01-02")
		td.XAxis.Data = append(td.XAxis.Data, today)
		record := records[today]
		switch request.Query.ResourceType {
		case "cpu":
			td.Series[0].Data = append(td.Series[0].Data, calcu.MillcoreToCore(record.ProdCPURequest, 3))
			td.Series[1].Data = append(td.Series[1].Data, calcu.MillcoreToCore(record.StagingCPURequest, 3))
			td.Series[2].Data = append(td.Series[2].Data, calcu.MillcoreToCore(record.TestCPURequest, 3))
			td.Series[3].Data = append(td.Series[3].Data, calcu.MillcoreToCore(record.DevCPURequest, 3))
		case "mem", "memory":
			td.Series[0].Data = append(td.Series[0].Data, calcu.ByteToGibibyte(record.ProdMemRequest, 3))
			td.Series[1].Data = append(td.Series[1].Data, calcu.ByteToGibibyte(record.StagingMemRequest, 3))
			td.Series[2].Data = append(td.Series[2].Data, calcu.ByteToGibibyte(record.TestMemRequest, 3))
			td.Series[3].Data = append(td.Series[3].Data, calcu.ByteToGibibyte(record.DevMemRequest, 3))
		default:
			td.Series[0].Data = append(td.Series[0].Data, float64(record.ProdPodsCount))
			td.Series[1].Data = append(td.Series[1].Data, float64(record.StagingPodsCount))
			td.Series[2].Data = append(td.Series[2].Data, float64(record.TestPodsCount))
			td.Series[3].Data = append(td.Series[3].Data, float64(record.DevPodsCount))
		}
	}
	return td, nil
}
