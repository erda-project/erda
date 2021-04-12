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

package apis

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda-infra/modcom/api"
	dicestructs "github.com/erda-project/erda/apistructs"
	block "github.com/erda-project/erda/modules/monitor/dashboard/chart-block"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/modules/pkg/mysql"
	"github.com/erda-project/erda/pkg/uuid"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (p *provider) creatOrgReportTask(obj *reportTask) interface{} {
	if len(obj.Scope) == 0 {
		obj.Scope = string(dicestructs.OrgScope)
	}
	obj.Enable = true
	var err error
	tx := p.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			p.Log.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	_, err = tx.systemBlock.Get(&block.DashboardBlockQuery{ID: obj.DashboardId})
	if err != nil && gorm.IsRecordNotFoundError(err) {
		return api.Errors.NotFound("dashboard block")
	}
	if err = tx.reportTask.Save(obj); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.AlreadyExists("report task")
		}
		return api.Errors.Internal(err)
	}
	// create pipeline and pipelineCron
	if err = p.createReportPipelineCron(obj); err != nil {
		return api.Errors.Internal(err)
	}
	if err = tx.reportTask.Save(obj); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.AlreadyExists("report task")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(obj)
}

func (p *provider) updateOrgReportTask(obj *reportTaskUpdate, params struct {
	ID uint64 `param:"id" validate:"required"`
}) interface{} {
	tx := p.db.Begin()
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			p.Log.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	report, err := tx.reportTask.Get(&reportTaskQuery{ID: &params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound(err)
		}
		return api.Errors.Internal(err)
	}
	report = editReportTaskFields(report, obj)
	if err = tx.reportTask.Save(report); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.AlreadyExists("report task")
		}
		return api.Errors.Internal(err)
	}

	return api.Success(true)
}

func editReportTaskFields(report *reportTask, update *reportTaskUpdate) *reportTask {
	if update.Name != nil {
		report.Name = *update.Name
	}
	if update.NotifyTarget != nil {
		report.NotifyTarget = update.NotifyTarget
	}
	if update.DashboardId != nil {
		report.DashboardId = *update.DashboardId
	}
	return report
}

func (p *provider) switchOrgReportTask(params struct {
	ID     uint64 `param:"id" validate:"required"`
	Enable bool   `query:"enable"`
}) interface{} {
	tx := p.db.Begin()
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			p.Log.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	report, err := tx.reportTask.Get(&reportTaskQuery{ID: &params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("report task")
		}
		return api.Errors.Internal(err)
	}
	cron, err := p.bdl.GetPipelineCron(report.PipelineCronId)
	if err != nil {
		return api.Errors.Internal(err)
	}
	if *cron.Enable && !params.Enable {
		_, err = p.bdl.StopPipelineCron(cron.ID)
		if err != nil {
			return api.Errors.Internal(err)
		}
	}
	if !*cron.Enable && params.Enable {
		_, err = p.bdl.StartPipelineCron(cron.ID)
		if err != nil {
			return api.Errors.Internal(err)
		}
	}
	report.Enable = params.Enable
	if err = tx.reportTask.Save(report); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.AlreadyExists("report task")
		}
		return api.Errors.Internal("failed to switch report task :", err)
	}

	return api.Success(true)
}

func (p *provider) listOrgReportTasks(r *http.Request, params struct {
	Scope    string `query:"scope" validate:""`
	ScopeID  string `query:"scopeId" validate:"required"`
	PageNo   int64  `query:"pageNo" validate:"gte=0"`
	PageSize int64  `query:"pageSize" validate:"gte=0,lte=100"`
	Type     string `query:"type" validate:""`
}) interface{} {
	if params.Scope == "" {
		params.Scope = dicestructs.OrgResource
	}
	if params.PageNo <= 0 {
		params.PageNo = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	query := reportTaskQuery{
		Scope:         params.Scope,
		ScopeID:       params.ScopeID,
		CreatedAtDesc: true,
	}
	if len(params.Type) > 0 {
		query.Type = params.Type
	}
	reports, total, err := p.db.reportTask.List(&query, params.PageSize, params.PageNo)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return api.Errors.Internal(err)
	}
	var reportDTOs []reportTaskDTO
	for _, obj := range reports {
		obj.NotifyTarget.NotifyGroup = p.getNotifyGroupRelByID(strconv.FormatUint(obj.NotifyTarget.GroupID, 10))
		reportDTO := reportTaskDTO{
			ID:           obj.ID,
			Name:         obj.Name,
			Scope:        obj.Scope,
			ScopeID:      obj.ScopeID,
			Type:         obj.Type,
			Enable:       obj.Enable,
			NotifyTarget: obj.NotifyTarget,
			DashboardId:  obj.DashboardId,
			CreatedAt:    utils.ConvertTimeToMS(obj.CreatedAt),
			UpdatedAt:    utils.ConvertTimeToMS(obj.UpdatedAt),
		}
		reportDTOs = append(reportDTOs, reportDTO)
	}
	return api.Success(&reportTaskResp{
		ReportTasks: reportDTOs,
		Total:       total,
	})
}

func (p *provider) runReportTaskAtOnce(params struct {
	ID uint64 `param:"id" validate:"required"`
}) interface{} {

	report, err := p.db.reportTask.Get(&reportTaskQuery{
		ID: &params.ID,
	})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("report task")
		}
		return api.Errors.Internal(err)
	}

	report.RunAtOnce = true
	report.Type = "" // cancel cron settings
	if err = p.createReportPipelineCron(report); err != nil {
		return api.Errors.Internal("failed to run report task at once :", err)
	}
	return api.Success(true)
}

func (p *provider) getOrgReportTask(params struct {
	ID uint64 `param:"id" validate:"required"`
}) interface{} {
	obj, err := p.db.reportTask.Get(&reportTaskQuery{
		ID:                    &params.ID,
		PreLoadDashboardBlock: true,
	})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("report task")
		}
		return api.Errors.Internal(err)
	}
	obj.NotifyTarget.NotifyGroup = p.getNotifyGroupRelByID(strconv.FormatUint(obj.NotifyTarget.GroupID, 10))
	return api.Success(&reportTaskDTO{
		ID:      obj.ID,
		Name:    obj.Name,
		Scope:   obj.Scope,
		ScopeID: obj.ScopeID,
		Type:    obj.Type,
		DashboardBlockTemplate: &block.DashboardBlockDTO{
			ID:         obj.DashboardBlock.ID,
			Name:       obj.DashboardBlock.Name,
			Desc:       obj.DashboardBlock.Desc,
			Scope:      obj.DashboardBlock.Scope,
			ScopeID:    obj.DashboardBlock.ScopeID,
			ViewConfig: obj.DashboardBlock.ViewConfig,
			DataConfig: obj.DashboardBlock.DataConfig,
			CreatedAt:  utils.ConvertTimeToMS(obj.CreatedAt),
			UpdatedAt:  utils.ConvertTimeToMS(obj.CreatedAt),
		},
		Enable:       obj.Enable,
		NotifyTarget: obj.NotifyTarget,
		CreatedAt:    utils.ConvertTimeToMS(obj.CreatedAt),
		UpdatedAt:    utils.ConvertTimeToMS(obj.UpdatedAt),
	})
}

func (p *provider) getNotifyGroupRelByID(groupID string) *dicestructs.NotifyGroup {
	if groupID == "" {
		return nil
	}
	notifyGroupsData, err := p.cmdb.QueryNotifyGroup([]string{groupID})
	if err != nil {
		logrus.Errorf("request cmdb: query notify group error: %s\n", err)
		return nil
	}
	if len(notifyGroupsData) > 0 {
		return notifyGroupsData[0]
	}

	return nil
}

func (p *provider) delOrgReportTask(params struct {
	ID uint64 `param:"id" validate:"required"`
}) interface{} {
	tx := p.db.Begin()
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			p.Log.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	query := reportTaskQuery{ID: &params.ID}
	report, err := tx.reportTask.Get(&query)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("report task")
		}
		return api.Errors.Internal("failed to get report task :", err)
	}
	err = p.stopAndDelPipelineCron(report)
	if err != nil {
		return api.Errors.Internal("report task failed to stop pipeline", err)
	}
	err = tx.reportTask.Del(&query)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("report task")
		}
		return api.Errors.Internal("failed to delete report task", err)
	}
	err = tx.reportHistory.Del(&reportHistoryQuery{TaskId: &report.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("report task")
		}
		return api.Errors.Internal(err)
	}

	if report != nil {
		return api.Success(map[string]interface{}{
			"name": report.Name,
		})
	}
	return api.Success(true)
}

func (p *provider) createReportPipelineCron(obj *reportTask) error {
	pipeline, err := p.generatePipeline(obj)
	if err != nil {
		return err
	}
	createResp, err := p.bdl.CreatePipeline(&pipeline)
	if err != nil {
		return err
	}
	if createResp.CronID != nil {
		obj.PipelineCronId = *createResp.CronID
	}
	return nil
}

// stop and delete pipelineCron , ignored error
func (p *provider) stopAndDelPipelineCron(obj *reportTask) error {
	if obj.PipelineCronId != 0 {
		_, err := p.bdl.StopPipelineCron(obj.PipelineCronId)
		_ = p.bdl.DeletePipelineCron(obj.PipelineCronId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) generatePipeline(r *reportTask) (pipeline dicestructs.PipelineCreateRequestV2, err error) {
	pipeline.PipelineYml, err = p.generatePipelineYml(r)
	if err != nil {
		return pipeline, err
	}
	pipeline.PipelineSource = dicestructs.PipelineSourceDice
	pipeline.PipelineYmlName = uuid.UUID() + ".yml"
	pipeline.ClusterName = p.Cfg.ClusterName
	pipeline.AutoRunAtOnce = r.RunAtOnce
	if r.Enable {
		pipeline.AutoStartCron = true
	} else {
		pipeline.AutoStartCron = false
	}

	return pipeline, nil
}

func (p *provider) generatePipelineYml(r *reportTask) (string, error) {
	pipelineYml := &dicestructs.PipelineYml{
		Version: p.Cfg.Pipeline.Version,
	}
	switch r.Type {
	case monthly:
		pipelineYml.Cron = p.Cfg.ReportCron.MonthlyCron
	case weekly:
		pipelineYml.Cron = p.Cfg.ReportCron.WeeklyCron
	case daily:
		pipelineYml.Cron = p.Cfg.ReportCron.DailyCron
	}
	org, err := p.bdl.GetOrg(r.ScopeID)
	if err != nil {
		return "", errors.Errorf("failed to generate pipeline yaml, can not get OrgName by OrgID:%v,(%+v)", r.ScopeID, err)
	}
	pipelineYml.Stages = [][]*dicestructs.PipelineYmlAction{{{
		Type:    p.Cfg.Pipeline.ActionType,
		Version: p.Cfg.Pipeline.ActionVersion,
		Params: map[string]interface{}{
			"report_id":   r.ID,
			"org_name":    org.Name,
			"domain_addr": fmt.Sprintf("%s://%s", p.Cfg.DiceProtocol, org.Domain),
		},
	}}}
	byteContent, err := yaml.Marshal(pipelineYml)
	if err != nil {
		return "", errors.Errorf("failed to generate pipeline yaml, pipelineYml:%+v, (%+v)", pipelineYml, err)
	}

	logrus.Debugf("[PipelineYml]: %s", string(byteContent))
	return string(byteContent), nil
}
