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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
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
	return dao.TestFileExtra{ManualTestFileExtraInfo: fileExtra.ManualTestFileExtraInfo}
}

func (svc *Service) GetFileRecord(id uint64) (*apistructs.TestFileRecord, error) {
	record, err := svc.db.GetRecord(id)
	if err != nil {
		return nil, apierrors.ErrGetFileRecord.InternalError(err)
	}
	return mapping(record), nil
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

func (svc *Service) ListFileRecordsByProject(req apistructs.ListTestFileRecordsRequest) ([]apistructs.TestFileRecord, []string, error) {
	if req.ProjectID == 0 {
		return nil, nil, apierrors.ErrListFileRecord.MissingParameter("projectID")
	}

	recordDtos, err := svc.db.ListRecordsByProject(req)
	if err != nil {
		return nil, nil, apierrors.ErrListFileRecord.InternalError(err)
	}

	records := make([]apistructs.TestFileRecord, 0)
	operators := make([]string, 0)
	for _, i := range recordDtos {
		records = append(records, *mapping(&i))
		operators = append(operators, i.OperatorID)
	}
	return records, operators, nil
}

func (svc *Service) GetFirstFileReady(actionType apistructs.FileActionType) (bool, *dao.TestFileRecord, error) {
	ok, record, err := svc.db.FirstFileReady(actionType)
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

func mapping(s *dao.TestFileRecord) *apistructs.TestFileRecord {
	return &apistructs.TestFileRecord{
		ID:          s.ID,
		FileName:    s.FileName,
		Description: s.Description,
		ProjectID:   s.ProjectID,
		ApiFileUUID: s.ApiFileUUID,
		Type:        s.Type,
		State:       s.State,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		OperatorID:  s.OperatorID,
	}
}
