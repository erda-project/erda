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

package apistructs

import (
	"time"
)

type TestFileRecord struct {
	ID          uint64          `json:"id"`
	FileName    string          `json:"name"`
	Description string          `json:"description"`
	ProjectID   uint64          `json:"projectID"`
	ApiFileUUID string          `json:"apiFileUUID"`
	Type        FileActionType  `json:"type"`
	State       FileRecordState `json:"state"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	OperatorID  string          `json:"operatorID"`
}

type TestFileRecordRequest struct {
	ID          uint64          `json:"id"`
	FileName    string          `json:"name"`
	ProjectID   uint64          `json:"projectID"`
	TestSetID   uint64          `json:"testSetID"`
	Description string          `json:"description"`
	ApiFileUUID string          `json:"apiFileUUID"`
	Type        FileActionType  `json:"type"`
	State       FileRecordState `json:"state"`
	Extra       TestFileExtra   `json:"extra"`
	IdentityInfo
}

type TestFileExtra struct {
	ManualTestFileExtraInfo *ManualTestFileExtraInfo `json:"manualTestExtraFileInfo,omitempty"`
}

type ManualTestFileExtraInfo struct {
	ImportRequest *TestCaseImportRequest `json:"importRequest,omitempty"`
	ExportRequest *TestCaseExportRequest `json:"exportRequest,omitempty"`
}

type FileRecordState string

type FileActionType string

const (
	FileRecordStatePending    FileRecordState = "pending"
	FileRecordStateProcessing FileRecordState = "processing"
	FileRecordStateSuccess    FileRecordState = "success"
	FileRecordStateFail       FileRecordState = "fail"
	FileActionTypeCopy        FileActionType  = "copy"
	FileActionTypeImport      FileActionType  = "import"
	FileActionTypeExport      FileActionType  = "export"
)

type ListTestFileRecordsRequest struct {
	ProjectID uint64           `json:"projectID"`
	Types     []FileActionType `json:"types"`
}

type GetTestFileRecordResponse struct {
	Header
	Data TestFileRecord
}

type ListTestFileRecordsResponse struct {
	Header
	Data []TestFileRecord
}
