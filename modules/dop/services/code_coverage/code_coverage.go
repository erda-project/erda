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
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

type CodeCoverage struct {
	db  CodeCoverageDBer
	bdl CodeCoverageBDLer
}

type Option func(*CodeCoverage)

func New(options ...Option) *CodeCoverage {
	c := &CodeCoverage{}
	for _, op := range options {
		op(c)
	}
	return c
}

func WithDBClient(db CodeCoverageDBer) Option {
	return func(c *CodeCoverage) {
		c.db = db
	}
}

func WithBundle(bdl CodeCoverageBDLer) Option {
	return func(c *CodeCoverage) {
		c.bdl = bdl
	}
}

// Start Start record
func (svc *CodeCoverage) Start(req apistructs.CodeCoverageStartRequest) error {
	// check permission
	if !req.IdentityInfo.IsInternalClient() {
		access, err := svc.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectID,
			Resource: "codeCoverage",
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return err
		}
		if !access.Access {
			return apierrors.ErrStartCodeCoverageExecRecord.AccessDenied()
		}
	}

	if err := svc.JudgeRunningRecordExist(req.ProjectID); err != nil {
		return err
	}

	record := dao.CodeCoverageExecRecord{
		ProjectID:     req.ProjectID,
		Status:        apistructs.RunningStatus,
		ReportStatus:  apistructs.RunningStatus,
		TimeBegin:     time.Now(),
		StartExecutor: req.UserID,
		TimeEnd:       time.Date(1000, 01, 01, 0, 0, 0, 0, time.UTC),
	}
	tx := svc.db.TxBegin()
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return err
	}

	jacocoAddress := GetJacocoAddr(record.ProjectID)
	if len(jacocoAddress) <= 0 {
		tx.Rollback()
		return fmt.Errorf("not find jaccoco application address")
	}

	// call jacoco start
	if err := svc.bdl.JacocoStart(jacocoAddress, &apistructs.JacocoRequest{
		ProjectID: record.ProjectID,
		PlanID:    record.ID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// End End record
func (svc *CodeCoverage) End(req apistructs.CodeCoverageUpdateRequest) error {
	record, err := svc.db.GetCodeCoverageByID(req.ID)
	if err != nil {
		return err
	}

	// check permission
	if !req.IdentityInfo.IsInternalClient() {
		access, err := svc.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  record.ProjectID,
			Resource: "codeCoverage",
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return err
		}
		if !access.Access {
			return apierrors.ErrEndCodeCoverageExecRecord.AccessDenied()
		}
	}

	if record.Status != apistructs.ReadyStatus {
		return errors.New("the pre status is not ready")
	}
	record.Status = apistructs.EndingStatus
	record.TimeEnd = time.Now()
	record.EndExecutor = req.UserID
	if err := svc.db.UpdateCodeCoverage(record); err != nil {
		return err
	}
	// call jacoco end
	jacocoAddress := GetJacocoAddr(record.ProjectID)
	if len(jacocoAddress) <= 0 {
		return fmt.Errorf("not find jaccoco application address")
	}
	return svc.bdl.JacocoEnd(jacocoAddress, &apistructs.JacocoRequest{
		ProjectID: record.ProjectID,
		PlanID:    record.ID,
	})
}

// Cancel Cancel record
func (svc *CodeCoverage) Cancel(req apistructs.CodeCoverageCancelRequest) error {
	// check permission
	if !req.IdentityInfo.IsInternalClient() {
		access, err := svc.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectID,
			Resource: "codeCoverage",
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return err
		}
		if !access.Access {
			return apierrors.ErrCancelCodeCoverageExecRecord.AccessDenied()
		}
	}
	record := dao.CodeCoverageExecRecord{
		Status:       apistructs.CancelStatus,
		ReportStatus: apistructs.CancelStatus,
		TimeEnd:      time.Now(),
	}

	return svc.db.CancelCodeCoverage(req.ProjectID, &record)
}

// ReadyCallBack Record ready callBack
func (svc *CodeCoverage) ReadyCallBack(req apistructs.CodeCoverageUpdateRequest) error {
	record, err := svc.db.GetCodeCoverageByID(req.ID)
	if err != nil {
		return err
	}
	if record.Status == apistructs.CancelStatus {
		return nil
	}

	if record.Status != apistructs.RunningStatus {
		return errors.New("the pre status is not running")
	}
	record.Status = apistructs.ReadyStatus
	record.Msg = req.Msg

	return svc.db.UpdateCodeCoverage(record)
}

// EndCallBack Record end callBack
func (svc *CodeCoverage) EndCallBack(req apistructs.CodeCoverageUpdateRequest) error {
	record, err := svc.db.GetCodeCoverageByID(req.ID)
	if err != nil {
		return err
	}

	status := apistructs.CodeCoverageExecStatus(req.Status)
	if status != apistructs.FailStatus && status != apistructs.SuccessStatus {
		return errors.New("the status is not fail or success")
	}
	if record.Status == apistructs.CancelStatus {
		return nil
	}

	project, err := svc.bdl.GetProject(record.ProjectID)
	if err != nil {
		return err
	}
	record.Status = status
	record.Msg = req.Msg
	if status == apistructs.FailStatus {
		record.ReportStatus = apistructs.FailStatus
	}

	if req.ReportXmlUUID != "" {
		f, err := svc.bdl.DownloadDiceFile(req.ReportXmlUUID)
		if err != nil {
			return err
		}
		defer f.Close()
		all, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		analyzeJson, coverage, err := getAnalyzeJson(project.ID, project.DisplayName, all)
		if err != nil {
			return err
		}
		record.ReportContent = analyzeJson
		record.Coverage = coverage
	}

	return svc.db.UpdateCodeCoverage(record)
}

// ReportCallBack Record report callBack
func (svc *CodeCoverage) ReportCallBack(req apistructs.CodeCoverageUpdateRequest) error {
	record, err := svc.db.GetCodeCoverageByID(req.ID)
	if err != nil {
		return err
	}
	reportStatus := apistructs.CodeCoverageExecStatus(req.Status)
	if reportStatus != apistructs.FailStatus && reportStatus != apistructs.SuccessStatus {
		return errors.New("the status is not fail or success")
	}
	if record.ReportStatus == apistructs.CancelStatus {
		return nil
	}

	record.ReportStatus = reportStatus
	record.ReportMsg = req.Msg
	record.ReportUrl = req.ReportTarUrl

	return svc.db.UpdateCodeCoverage(record)
}

// ListCodeCoverageRecord list code coverage record
func (svc *CodeCoverage) ListCodeCoverageRecord(req apistructs.CodeCoverageListRequest) (data apistructs.CodeCoverageExecRecordData, err error) {
	var (
		records []dao.CodeCoverageExecRecordShort
		list    []apistructs.CodeCoverageExecRecordDto
		total   uint64
	)
	records, total, err = svc.db.ListCodeCoverage(req)
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
	record, err := svc.db.GetCodeCoverageByID(id)
	if err != nil {
		return nil, err
	}
	return record.Covert(), nil
}

// JudgeRunningRecordExist Judge running record exist
func (svc *CodeCoverage) JudgeRunningRecordExist(projectID uint64) error {
	records, err := svc.db.ListCodeCoverageByStatus(projectID, apistructs.WorkingStatus)
	if err != nil {
		return err
	}
	if len(records) > 0 {
		return errors.New("there is already running record")
	}
	return nil
}

func (svc *CodeCoverage) JudgeCanEnd(projectID uint64) (bool, error) {
	records, err := svc.db.ListCodeCoverageByStatus(projectID, []apistructs.CodeCoverageExecStatus{apistructs.ReadyStatus})
	if err != nil {
		return false, err
	}
	if len(records) > 0 {
		return true, nil
	}

	return false, nil
}

func GetJacocoAddr(projectID uint64) string {
	return conf.JacocoAddr()[strconv.FormatUint(projectID, 10)]
}

type CodeCoverageDBer interface {
	CreateCodeCoverage(record *dao.CodeCoverageExecRecord) error
	UpdateCodeCoverage(record *dao.CodeCoverageExecRecord) error
	GetCodeCoverageByID(id uint64) (*dao.CodeCoverageExecRecord, error)
	CancelCodeCoverage(uint64, *dao.CodeCoverageExecRecord) error
	ListCodeCoverageByStatus(projectID uint64, status []apistructs.CodeCoverageExecStatus) (records []dao.CodeCoverageExecRecord, err error)
	ListCodeCoverage(req apistructs.CodeCoverageListRequest) (records []dao.CodeCoverageExecRecordShort, total uint64, err error)
	TxBegin() *gorm.DB
}

type CodeCoverageBDLer interface {
	CheckPermission(req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error)
	JacocoStart(addr string, req *apistructs.JacocoRequest) error
	JacocoEnd(addr string, req *apistructs.JacocoRequest) error
	DownloadDiceFile(uuid string) (io.ReadCloser, error)
	GetProject(id uint64) (*apistructs.ProjectDTO, error)
}
