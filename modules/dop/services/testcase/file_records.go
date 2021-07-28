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

package testcase

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/i18n"
)

func (svc *Service) CreateFileRecord(req apistructs.TestFileRecordRequest) (uint64, error) {
	record := &dao.TestFileRecord{
		FileName:    req.FileName,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		Type:        req.Type,
		State:       req.State,
		ApiFileUUID: req.ApiFileUUID,
		OperatorID:  req.UserID,
		Extra:       convertTestFileExtra(req.Extra),
	}

	if err := svc.db.CreateRecord(record); err != nil {
		return 0, apierrors.ErrCreateFileRecord.InternalError(err)
	}
	return record.ID, nil
}

func convertTestFileExtra(fileExtra apistructs.TestFileExtra) dao.TestFileExtra {
	return dao.TestFileExtra{
		ManualTestFileExtraInfo:    fileExtra.ManualTestFileExtraInfo,
		AutotestSpaceFileExtraInfo: fileExtra.AutotestSpaceFileExtraInfo,
	}
}

func (svc *Service) GetFileRecord(id uint64, locale string) (*apistructs.TestFileRecord, error) {
	record, err := svc.db.GetRecord(id)
	if err != nil {
		return nil, apierrors.ErrGetFileRecord.InternalError(err)
	}
	l := svc.bdl.GetLocale(locale)
	project := l.Get(i18n.I18nKeyProjectName)
	testSet := l.Get(i18n.I18nKeyCaseSetName)
	return mapping(record, project, testSet), nil
}

func (svc *Service) UpdateFileRecord(req apistructs.TestFileRecordRequest) error {
	r, err := svc.db.GetRecord(req.ID)
	if err != nil {
		return apierrors.ErrUpdateFileRecord.InternalError(err)
	}

	if req.ApiFileUUID != "" {
		r.ApiFileUUID = req.ApiFileUUID
	}
	if req.State != "" {
		r.State = req.State
	}
	return svc.db.UpdateRecord(r)
}

func (svc *Service) ListFileRecordsByProject(req apistructs.ListTestFileRecordsRequest) ([]apistructs.TestFileRecord, []string, map[string]int, error) {
	if req.ProjectID == 0 {
		return nil, nil, nil, apierrors.ErrListFileRecord.MissingParameter("projectID")
	}

	recordDtos, count, err := svc.db.ListRecordsByProject(req)
	if err != nil {
		return nil, nil, nil, apierrors.ErrListFileRecord.InternalError(err)
	}

	records := make([]apistructs.TestFileRecord, 0)

	l := svc.bdl.GetLocale(req.Locale)
	project := l.Get(i18n.I18nKeyProjectName)
	testSet := l.Get(i18n.I18nKeyCaseSetName)

	operators := make([]string, 0)
	for _, i := range recordDtos {
		records = append(records, *mapping(&i, project, testSet))
		operators = append(operators, i.OperatorID)
	}
	return records, operators, count, nil
}

func (svc *Service) GetFirstFileReady(actionType ...apistructs.FileActionType) (bool, *dao.TestFileRecord, error) {
	ok, record, err := svc.db.FirstFileReady(actionType...)
	if err != nil || !ok {
		return false, nil, err
	}
	return true, record, nil
}

func (svc *Service) BatchClearProcessingRecords() error {
	return svc.db.BatchUpdateRecords()
}

func (svc *Service) DeleteRecordApiFilesByTime(t time.Time) error {
	UUIDs, err := svc.db.DeleteFileRecordByTime(t)
	if err != nil {
		return err
	}
	for _, UUID := range UUIDs {
		if UUID.ApiFileUUID == "" {
			continue
		}
		if err := svc.bdl.DeleteDiceFile(UUID.ApiFileUUID); err != nil {
			return err
		}
	}
	return nil
}

func mapping(s *dao.TestFileRecord, project, testSet string) *apistructs.TestFileRecord {
	record := &apistructs.TestFileRecord{
		ID:          s.ID,
		FileName:    s.FileName,
		ProjectID:   s.ProjectID,
		ApiFileUUID: s.ApiFileUUID,
		TestSetID: func() uint64 {
			if info := s.Extra.ManualTestFileExtraInfo; info != nil {
				return info.TestSetID
			}
			return 0
		}(),
		Description: s.Description,
		Type:        s.Type,
		State:       s.State,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		OperatorID:  s.OperatorID,
	}

	if record.Type == apistructs.FileActionTypeImport || record.Type == apistructs.FileActionTypeExport {
		record.Description = fmt.Sprintf("%v ID: %v, %v ID: %v", project, record.ProjectID, testSet, record.TestSetID)
	}
	return record
}
