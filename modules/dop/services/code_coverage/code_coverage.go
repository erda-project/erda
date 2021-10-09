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

package code_coverage

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
)

type CodeCoverage struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

type Option func(*CodeCoverage)

func New(options ...Option) *CodeCoverage {
	c := &CodeCoverage{}
	for _, op := range options {
		op(c)
	}
	return c
}

func WithDBClient(db *dao.DBClient) Option {
	return func(c *CodeCoverage) {
		c.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(c *CodeCoverage) {
		c.bdl = bdl
	}
}

// Start Start record
func (svc *CodeCoverage) Start(req apistructs.CodeCoverageStartRequest) error {
	now := time.Now()
	record := dao.CodeCoverageExecRecord{
		ProjectID:     req.ProjectID,
		Status:        apistructs.RunningStatus,
		TimeBegin:     &now,
		StartExecutor: req.UserID,
	}
	if err := svc.db.Create(&record).Error; err != nil {
		return err
	}
	// call jacoco start
	return svc.bdl.JacocoStart(&apistructs.JacocoRequest{
		ProjectID: record.ProjectID,
		PlanID:    record.ID,
	})
}

// End End record
func (svc *CodeCoverage) End(req apistructs.CodeCoverageUpdateRequest) error {
	var record dao.CodeCoverageExecRecord
	if err := svc.db.Model(&dao.CodeCoverageExecRecord{}).First(&record, req.ID).Error; err != nil {
		return err
	}
	if record.Status != apistructs.ReadyStatus {
		return errors.New("the pre status is not ready")
	}
	record.Status = apistructs.EndingStatus
	now := time.Now()
	record.TimeEnd = &now
	record.EndExecutor = req.UserID
	if err := svc.db.Save(&record).Error; err != nil {
		return err
	}
	// call jacoco end
	return svc.bdl.JacocoEnd(&apistructs.JacocoRequest{
		ProjectID: record.ProjectID,
		PlanID:    record.ID,
	})
}

// ReadyCallBack Record ready callBack
func (svc *CodeCoverage) ReadyCallBack(req apistructs.CodeCoverageUpdateRequest) error {
	var record dao.CodeCoverageExecRecord
	if err := svc.db.Model(&dao.CodeCoverageExecRecord{}).First(&record, req.ID).Error; err != nil {
		return err
	}
	status := apistructs.CodeCoverageExecStatus(req.Status)
	if status != apistructs.ReadyStatus {
		return errors.New("the status is not ready")
	}

	if record.Status != apistructs.RunningStatus {
		return errors.New("the pre status is not running")
	}
	record.Status = status
	record.Msg = req.Msg

	return svc.db.Save(&record).Error
}

// EndCallBack Record end callBack
func (svc *CodeCoverage) EndCallBack(req apistructs.CodeCoverageUpdateRequest) error {
	var record dao.CodeCoverageExecRecord
	if err := svc.db.Model(&dao.CodeCoverageExecRecord{}).First(&record, req.ID).Error; err != nil {
		return err
	}
	status := apistructs.CodeCoverageExecStatus(req.Status)
	if status != apistructs.FailStatus && status != apistructs.SuccessStatus {
		return errors.New("the status is not fail or success")
	}

	record.Status = status
	record.Msg = req.Msg
	// upload report_tar
	if req.ReportTar != nil {
		f, err := req.ReportTar.Open()
		if err != nil {
			return err
		}
		uploadReq := apistructs.FileUploadRequest{
			FileNameWithExt: req.ReportTar.Filename,
			ByteSize:        req.ReportTar.Size,
			FileReader:      f,
			From:            "Autotest space",
			IsPublic:        true,
			ExpiredAt:       nil,
		}
		file, err := svc.bdl.UploadFile(uploadReq)
		if err != nil {
			return err
		}
		record.ReportUrl = file.DownloadURL
	}
	// deal report_xml
	if req.ReportXml != nil {
		f, err := req.ReportXml.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		all, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		analyzeJson, coverage := getAnalyzeJson(all)
		record.ReportContent = analyzeJson
		record.Coverage = coverage
	}

	return svc.db.Save(&record).Error
}

// ListCodeCoverageRecord list code coverage record
func (svc *CodeCoverage) ListCodeCoverageRecord(req apistructs.CodeCoverageListRequest) (data apistructs.CodeCoverageExecRecordData, err error) {
	var (
		records []dao.CodeCoverageExecRecordShort
		list    []apistructs.CodeCoverageExecRecordDto
		total   uint64
	)

	offset := (req.PageNo - 1) * req.PageSize
	db := svc.db.Model(&dao.CodeCoverageExecRecordShort{}).
		Where("project_id = ?", req.ProjectID)

	if req.Statuses != nil {
		db = db.Where("status in ?", req.Statuses)
	}
	if req.TimeBegin != "" {
		db = db.Where("time_begin >= ?", req.TimeBegin)
	}
	if req.TimeEnd != "" {
		db = db.Where("time_begin <= ?", req.TimeEnd)
	}

	if req.Asc {
		db = db.Order("id ASC")
	} else {
		db = db.Order("id DESC")
	}

	err = db.Offset(offset).Limit(req.PageSize).
		Find(&records).
		Offset(0).Limit(-1).Count(&total).Error
	if err != nil {
		return
	}
	for _, v := range records {
		list = append(list, v.Covert())
	}
	return apistructs.CodeCoverageExecRecordData{
		Total: total,
		List:  list,
	}, nil
}

// GetCodeCoverageRecord Get code coverage record
func (svc *CodeCoverage) GetCodeCoverageRecord(id uint64) (*apistructs.CodeCoverageExecRecordDto, error) {
	var record dao.CodeCoverageExecRecord
	if err := svc.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return record.Covert(), nil
}

// JudgeRunningRecordExist Judge running record exist
func (svc *CodeCoverage) JudgeRunningRecordExist(projectID uint64) error {
	var records []dao.CodeCoverageExecRecord
	if err := svc.db.Where("project_id = ?", projectID).Where("status IN (?)", apistructs.WorkingStatus).Find(&records).Error; err != nil {
		return err
	}
	if len(records) > 0 {
		return errors.New("there is already running record")
	}
	return nil
}

func (svc *CodeCoverage) JudgeCanEnd(projectID uint64) (bool, error) {
	var records []dao.CodeCoverageExecRecord
	if err := svc.db.Where("project_id = ?", projectID).Where("status IN (?)", apistructs.ReadyStatus).Find(&records).Error; err != nil {
		return false, err
	}
	if len(records) > 0 {
		return true, nil
	}

	return false, fmt.Errorf("not find ready status record")
}
