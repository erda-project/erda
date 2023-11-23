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

package cr

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

const MAX_FILE_CHANGES_CHAR_SIZE = 9 * 1000

type CodeReviewer interface {
	CodeReview() string
}

type FileCodeReviewer interface {
	CodeReviewer
	GetFileName() string
}

type AIInvoker interface {
	ConstructAIRequest() openai.ChatCompletionRequest
}

func NewCodeReviewer(req models.AICodeReviewNoteRequest, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo) (cr CodeReviewer, err error) {
	switch req.Type {
	case models.AICodeReviewTypeMR:
		cr = newMRReviewer(repo, mr)
		return
	case models.AICodeReviewTypeMRFile:
		if req.FileRelated == nil || req.FileRelated.NewFilePath == "" {
			return nil, fmt.Errorf("no file specified")
		}
		diffFile := getDiffFileFromMR(repo, mr, req.FileRelated.NewFilePath)
		if diffFile == nil {
			return nil, fmt.Errorf("file not found")
		}
		cr = newFileReviewer(diffFile, convertToModelUser(mr.AuthorUser), mr)
		return
	case models.AICodeReviewTypeMRCodeSnippet:
		if req.CodeSnippetRelated == nil || req.CodeSnippetRelated.SelectedCode == "" {
			return nil, fmt.Errorf("no code selected")
		}
		if req.CodeSnippetRelated.CodeLanguage == "" && req.CodeSnippetRelated.NewFilePath != "" {
			ext := strings.TrimPrefix(filepath.Ext(req.CodeSnippetRelated.NewFilePath), ".")
			req.CodeSnippetRelated.CodeLanguage = ext
		}
		cr = newSnippetCodeReviewer(req.CodeSnippetRelated.CodeLanguage, req.CodeSnippetRelated.SelectedCode, false, convertToModelUser(mr.AuthorUser))
		return
	default:
		return nil, fmt.Errorf("unknown code review type")
	}
}
