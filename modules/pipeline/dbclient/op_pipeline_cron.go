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

package dbclient

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// return: result, total, nil
func (client *Client) PagingPipelineCron(req apistructs.PipelineCronPagingRequest) ([]spec.PipelineCron, int64, error) {
	result := make([]spec.PipelineCron, 0)
	var total int64 = 0

	if req.PageNo < 0 {
		return nil, -1, errors.Errorf("invalid pageNo: %d", req.PageNo)
	}
	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize < 0 {
		return nil, -1, errors.Errorf("invalid pageSize: %d", req.PageSize)
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	limitSQL := client.Desc("id")
	// source
	// 必须指定 sources 或 allSources
	if !req.AllSources && len(req.Sources) == 0 {
		return nil, -1, errors.Errorf("source is empty")
	}
	if !req.AllSources {
		limitSQL.In("pipeline_source", req.Sources)
	}
	// ymlName
	if len(req.YmlNames) == 0 {
		return nil, -1, errors.Errorf("ymlName is empty")
	}
	limitSQL.In("pipeline_yml_name", req.YmlNames)

	// total
	totalSQL := *limitSQL
	var totalIDs []spec.PipelineCron
	if err := totalSQL.Find(&totalIDs); err != nil {
		return nil, -1, err
	}
	total = int64(len(totalIDs))

	// LIMIT ${pageSize} OFFSET ${pageNo}
	var limitIDs []spec.PipelineCron
	if err := limitSQL.Limit(req.PageSize, (req.PageNo-1)*req.PageSize).Find(&limitIDs); err != nil {
		return nil, -1, err
	}
	// get pipeline by header
	var ids []uint64
	for i := range limitIDs {
		ids = append(ids, limitIDs[i].ID)
	}
	if err := client.In("id", ids).Desc("id").Find(&result); err != nil {
		return nil, -1, err
	}

	return result, total, nil
}

func (client *Client) ListPipelineCronsByApplicationID(applicationID uint64) (crons []spec.PipelineCron, err error) {
	if err = client.Find(&crons, spec.PipelineCron{ApplicationID: applicationID}); err != nil {
		return nil, err
	}
	return crons, nil
}

func (client *Client) ListAllPipelineCrons() (crons []spec.PipelineCron, err error) {
	if err = client.Find(&crons); err != nil {
		return nil, err
	}
	return crons, nil
}

func (client *Client) ListPipelineCrons(enable *bool, ops ...SessionOption) ([]spec.PipelineCron, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	if enable != nil {
		session.Where("enable=?", *enable)
	}
	var crons []spec.PipelineCron
	err := session.Find(&crons)
	return crons, err
}

func (client *Client) GetPipelineCron(id interface{}) (cron spec.PipelineCron, err error) {
	defer func() {
		err = errors.Wrapf(err, "failed to get pipeline cron by id [%v]", id)
	}()
	found, err := client.ID(id).Get(&cron)
	if err != nil {
		return spec.PipelineCron{}, err
	}
	if !found {
		return spec.PipelineCron{}, errors.New("not found")
	}
	return cron, nil
}

func (client *Client) CheckExistPipelineCronByApplicationBranchYmlName(applicationID uint64, branch string, pipelineYmlName string) (bool, spec.PipelineCron, error) {
	var result = spec.PipelineCron{
		ApplicationID:   applicationID,
		Branch:          branch,
		PipelineYmlName: pipelineYmlName,
	}
	exist, err := client.Get(&result)
	if err != nil {
		return false, spec.PipelineCron{}, err
	}
	return exist, result, nil
}

func (client *Client) CreatePipelineCron(cron *spec.PipelineCron) error {
	_, err := client.InsertOne(cron)
	return errors.Wrapf(err, "failed to create pipeline cron, applicationID [%d], branch [%s], expr [%s], enable [%v]", cron.ApplicationID, cron.Branch, cron.CronExpr, cron.Enable)
}

func (client *Client) DeletePipelineCron(id interface{}) error {
	if _, err := client.ID(id).Delete(&spec.PipelineCron{}); err != nil {
		return errors.Errorf("failed to delete pipeline cron, id: %d, err: %v", id, err)
	}
	return nil
}

//更新cron的enable = false，cronExpr = new_.CronExpr
func (client *Client) DisablePipelineCron(new_ *spec.PipelineCron) (cronID uint64, err error) {

	var disable = false
	var updateCron = &spec.PipelineCron{}
	var columns = []string{spec.PipelineCronCronExpr, spec.PipelineCronEnable}
	// 寻找 v1
	queryV1 := &spec.PipelineCron{
		ApplicationID:   new_.ApplicationID,
		Branch:          new_.Branch,
		PipelineYmlName: new_.PipelineYmlName,
	}
	v1Exist, err := client.Get(queryV1)
	if err != nil {
		return 0, err
	}
	if queryV1.Extra.Version == "v2" {
		v1Exist = false
	}
	//只更新enable为false
	if v1Exist {
		updateCron.Enable = &disable
		updateCron.ID = queryV1.ID
		updateCron.CronExpr = new_.CronExpr
		return updateCron.ID, client.UpdatePipelineCronWillUseDefault(updateCron.ID, updateCron, columns)
	}

	//------------------------ v2
	//updateCron = &spec.PipelineCron{}
	// 寻找 v2
	queryV2 := &spec.PipelineCron{
		PipelineSource:  new_.PipelineSource,
		PipelineYmlName: new_.PipelineYmlName,
	}
	v2Exist, err := client.Get(queryV2)
	if err != nil {
		return 0, err
	}
	//只更新enable为false
	if v2Exist {
		updateCron.Enable = &disable
		updateCron.ID = queryV2.ID
		updateCron.CronExpr = new_.CronExpr
		return updateCron.ID, client.UpdatePipelineCronWillUseDefault(updateCron.ID, updateCron, columns)
	}

	return 0, nil
}

func (client *Client) UpdatePipelineCron(id interface{}, cron *spec.PipelineCron) error {
	_, err := client.ID(id).Update(cron)
	return errors.Wrapf(err, "failed to update pipeline cron, id [%v]", id)
}

//更新cron，但是假如传入的cron有默认值，那么数据库会被更新成默认值
func (client *Client) UpdatePipelineCronWillUseDefault(id interface{}, cron *spec.PipelineCron, columns []string) error {
	_, err := client.ID(id).Cols(columns...).Update(cron)
	return errors.Wrapf(err, "failed to update pipeline cron, id [%v]", id)
}

func (client *Client) InsertOrUpdatePipelineCron(new_ *spec.PipelineCron) error {

	// 寻找 v1
	queryV1 := &spec.PipelineCron{
		ApplicationID:   new_.ApplicationID,
		Branch:          new_.Branch,
		PipelineYmlName: new_.PipelineYmlName,
	}
	v1Exist, err := client.Get(queryV1)
	if err != nil {
		return err
	}
	if queryV1.Extra.Version == "v2" {
		v1Exist = false
	}
	if v1Exist {
		new_.ID = queryV1.ID
		new_.Enable = queryV1.Enable
		return client.UpdatePipelineCron(new_.ID, new_)
	}

	// 寻找 v2
	queryV2 := &spec.PipelineCron{
		PipelineSource:  new_.PipelineSource,
		PipelineYmlName: new_.PipelineYmlName,
	}
	v2Exist, err := client.Get(queryV2)
	if err != nil {
		return err
	}
	if v2Exist {
		new_.ID = queryV2.ID
		new_.Enable = queryV2.Enable
		return client.UpdatePipelineCron(new_.ID, new_)
	}
	return client.CreatePipelineCron(new_)
}
