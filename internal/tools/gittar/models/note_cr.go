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

package models

type AICodeReviewNoteRequest struct {
	NoteLocation NoteRequest `json:"noteLocation"`

	Type AICodeReviewType `json:"type,omitempty"`

	FileRelated        *AICodeReviewRequestForFile        `json:"fileRelated,omitempty"`
	CodeSnippetRelated *AICodeReviewRequestForCodeSnippet `json:"codeSnippetRelated,omitempty"`
}

type AICodeReviewType string

var (
	AICodeReviewTypeMR            AICodeReviewType = "MR"
	AICodeReviewTypeMRFile        AICodeReviewType = "MR_FILE"
	AICodeReviewTypeMRCodeSnippet AICodeReviewType = "MR_CODE_SNIPPET"
)

type AICodeReviewRequestForMR struct{}
type AICodeReviewRequestForFile struct {
	NewFilePath string `json:"newFilePath,omitempty"`
}
type AICodeReviewRequestForCodeSnippet struct {
	NewFilePath  string `json:"newFilePath,omitempty"`
	CodeLanguage string `json:"codeLanguage,omitempty"`
	SelectedCode string `json:"selectedCode,omitempty"`
}
