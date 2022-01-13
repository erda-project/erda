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

package apistructs

import (
	"net/url"
	"strconv"
	"time"
)

type TestFileRecord struct {
	ID                 uint64          `json:"id"`
	FileName           string          `json:"name"`
	Description        string          `json:"description"`
	ProjectName        string          `json:"projectName"`
	ProjectDisplayName string          `json:"projectDisplayName"`
	OrgID              uint64          `json:"orgID"`
	ProjectID          uint64          `json:"projectID"`
	TestSetID          uint64          `json:"testSetID"`
	ApiFileUUID        string          `json:"apiFileUUID"`
	SpaceID            uint64          `json:"spaceID"`
	Type               FileActionType  `json:"type"`
	State              FileRecordState `json:"state"`
	ErrorInfo          string          `json:"errorInfo"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
	OperatorID         string          `json:"operatorID"`
}

type TestFileRecordRequest struct {
	ID          uint64          `json:"id"`
	FileName    string          `json:"name"`
	OrgID       uint64          `json:"orgID"`
	ProjectID   uint64          `json:"projectID"`
	SpaceID     uint64          `json:"spaceID"`
	Description string          `json:"description"`
	ApiFileUUID string          `json:"apiFileUUID"`
	Type        FileActionType  `json:"type"`
	State       FileRecordState `json:"state"`
	Extra       TestFileExtra   `json:"extra"`
	ErrorInfo   error           `json:"errorInfo"`
	IdentityInfo
}

type TestFileExtra struct {
	ManualTestFileExtraInfo       *ManualTestFileExtraInfo       `json:"manualTestExtraFileInfo,omitempty"`
	AutotestSpaceFileExtraInfo    *AutoTestSpaceFileExtraInfo    `json:"autotestSpaceFileExtraInfo,omitempty"`
	AutotestSceneSetFileExtraInfo *AutoTestSceneSetFileExtraInfo `json:"autotestSceneSetFileExtraInfo,omitempty"`
	ProjectTemplateFileExtraInfo  *ProjectTemplateFileExtraInfo  `json:"projectTemplateFileExtraInfo,omitempty"`
}

const TestFileRecordErrorMaxLength = 2048

type ManualTestFileExtraInfo struct {
	TestSetID     uint64                   `json:"testSetID,omitempty"`
	ImportRequest *TestCaseImportRequest   `json:"importRequest,omitempty"`
	ExportRequest *TestCaseExportRequest   `json:"exportRequest,omitempty"`
	CopyRequest   *TestSetCopyAsyncRequest `json:"copyRequest,omitempty"`
}

type AutoTestSpaceFileExtraInfo struct {
	ImportRequest *AutoTestSpaceImportRequest `json:"importRequest,omitempty"`
	ExportRequest *AutoTestSpaceExportRequest `json:"exportRequest,omitempty"`
}

type AutoTestSceneSetFileExtraInfo struct {
	ExportRequest *AutoTestSceneSetExportRequest `json:"exportRequest,omitempty"`
	ImportRequest *AutoTestSceneSetImportRequest `json:"importRequest"`
}

type ProjectTemplateFileExtraInfo struct {
	ExportRequest *ExportProjectTemplateRequest `json:"exportRequest,omitempty"`
	ImportRequest *ImportProjectTemplateRequest `json:"importRequest,omitempty"`
}

type FileRecordState string

type FileActionType string

const (
	FileRecordStatePending       FileRecordState = "pending"
	FileRecordStateProcessing    FileRecordState = "processing"
	FileRecordStateSuccess       FileRecordState = "success"
	FileRecordStateFail          FileRecordState = "fail"
	FileActionTypeCopy           FileActionType  = "copy"
	FileActionTypeImport         FileActionType  = "import"
	FileActionTypeExport         FileActionType  = "export"
	FileSpaceActionTypeExport    FileActionType  = "spaceExport"
	FileSpaceActionTypeImport    FileActionType  = "spaceImport"
	FileSceneSetActionTypeExport FileActionType  = "sceneSetExport"
	FileSceneSetActionTypeImport FileActionType  = "sceneSetImport"
	FileProjectTemplateExport    FileActionType  = "projectTemplateExport"
	FileProjectTemplateImport    FileActionType  = "projectTemplateImport"
)

type ListTestFileRecordsRequest struct {
	ProjectID          uint64           `json:"projectID"`
	ProjectIDs         []uint64         `json:"-"`
	ProjectName        string           `json:"projectName"`
	ProjectDisplayName string           `json:"projectDisplayName"`
	OrgID              uint64           `json:"orgID"`
	SpaceID            uint64           `json:"spaceID"`
	Types              []FileActionType `json:"types"`
	Locale             string           `json:"locale"`
	PageNo             int              `json:"pageNo"`
	PageSize           int              `json:"pageSize"`
	Asc                bool             `json:"asc"`

	IdentityInfo
}

func (r ListTestFileRecordsRequest) ConvertToQueryParams() url.Values {
	values := make(url.Values)
	if r.ProjectID != 0 {
		values.Add("projectID", strconv.FormatInt(int64(r.ProjectID), 10))
	}
	if r.Locale != "" {
		values.Add("locale", r.Locale)
	}
	for _, fileType := range r.Types {
		values.Add("types", string(fileType))
	}
	return values
}

type GetTestFileRecordResponse struct {
	Header
	Data TestFileRecord
}

type ListTestFileRecordsResponse struct {
	Header
	Data *ListTestFileRecordsResponseData
}

type ListTestFileRecordsResponseData struct {
	Counter map[string]int   `json:"counter"`
	List    []TestFileRecord `json:"list"`
	Total   int              `json:"total"`
}
