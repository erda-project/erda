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

import (
	"fmt"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

type CodeReviewer interface {
	CodeReview(i18n i18n.Translator, lang i18n.LanguageCodes, aiSessionID string) string
}

type FileCodeReviewer interface {
	CodeReviewer
	GetFileName() string
}

type CodeReviewerCreateFunc func(req AICodeReviewNoteRequest, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, user *User) (CodeReviewer, error)

var CodeReviewerFactory = map[AICodeReviewType]CodeReviewerCreateFunc{}

func RegisterCodeReviewer(t AICodeReviewType, f CodeReviewerCreateFunc) {
	CodeReviewerFactory[t] = f
}

type AICodeReviewNoteRequest struct {
	NoteLocation NoteRequest `json:"noteLocation"`

	Type               AICodeReviewType                   `json:"type,omitempty"`
	CodeSnippetRelated *AICodeReviewRequestForCodeSnippet `json:"codeSnippetRelated,omitempty"`
}

type AICodeReviewType string

var (
	AICodeReviewTypeMR            AICodeReviewType = "MR"
	AICodeReviewTypeMRFile        AICodeReviewType = "MR_FILE"
	AICodeReviewTypeMRCodeSnippet AICodeReviewType = "MR_CODE_SNIPPET"
)

type AICodeReviewRequestForMR struct{}
type AICodeReviewRequestForCodeSnippet struct {
	CodeLanguage string `json:"codeLanguage,omitempty"` // if empty, will parse by newFilePath
	SelectedCode string `json:"selectedCode,omitempty"`
}

func NewCodeReviewer(req AICodeReviewNoteRequest, repo *gitmodule.Repository, user *User, mr *apistructs.MergeRequestInfo) (cr CodeReviewer, err error) {
	f, ok := CodeReviewerFactory[req.Type]
	if !ok {
		return nil, fmt.Errorf("unknown code review type: %s", req.Type)
	}
	return f(req, repo, mr, user)
}
