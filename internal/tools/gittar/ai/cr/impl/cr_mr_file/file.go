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

package cr_mr_file

import (
	"bytes"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mohae/deepcopy"
	"github.com/sashabaranov/go-openai"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/aiutil"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/i18nutil"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/mdutil"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/mrutil"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

type OneChangedFile struct {
	FileName     string
	CodeLanguage string
	FileContent  string
	Truncated    bool

	user *models.User
}

func NewFileReviewer(filePath string, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, user *models.User) (models.FileCodeReviewer, error) {
	if filePath == "" {
		return nil, fmt.Errorf("no file specified")
	}
	fileContent, truncated, err := mrutil.GetFileContent(repo, mr, filePath)
	if err != nil {
		return nil, err
	}

	fr := OneChangedFile{
		FileName:     filePath,
		CodeLanguage: strings.TrimPrefix(filepath.Ext(filePath), "."),
		FileContent:  fileContent,
		Truncated:    truncated,

		user: user,
	}
	return &fr, nil
}

func init() {
	models.RegisterCodeReviewer(models.AICodeReviewTypeMRFile, func(req models.AICodeReviewNoteRequest, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, user *models.User) (models.CodeReviewer, error) {
		return NewFileReviewer(req.NoteLocation.NewPath, repo, mr, user)
	})
}

func (r *OneChangedFile) GetFileName() string {
	return r.FileName
}

// CodeReview for file level, invoke once with all code snippets.
func (r *OneChangedFile) CodeReview(i18n i18n.Translator, lang i18n.LanguageCodes) string {
	// invoke ai
	result := aiutil.InvokeAI(r.constructAIRequest(i18n, lang), r.user)

	// truncate
	if r.Truncated {
		// calculate how many lines of file content
		lines := strings.Split(r.FileContent, "\n")
		lineCount := len(lines)
		truncatedTip := fmt.Sprintf(i18n.Text(lang, models.I18nKeyTemplateMrAICrFileContentMaxLimit), lineCount)
		truncatedTip = mdutil.MakeRef(mdutil.MakeItalic(truncatedTip))
		result = truncatedTip + "\n\n" + result
	}

	return result
}

type (
	FileReviewResult struct {
		Result []FileReviewResultItem `json:"fileReviewResult,omitempty"`
	}
	FileReviewResultItem struct {
		SnippetIndex int    `json:"snippetIndex,omitempty"`
		RiskLevel    string `json:"riskLevel,omitempty"`
		Details      string `json:"details,omitempty"`
	}
)

//go:embed prompt.yaml
var promptYaml string

type PromptStruct = struct {
	Messages []openai.ChatCompletionMessage `yaml:"messages"`
}

var promptStruct PromptStruct

func init() {
	if err := yaml.Unmarshal([]byte(promptYaml), &promptStruct); err != nil {
		panic(err)
	}
}

func (r *OneChangedFile) constructAIRequest(i18n i18n.Translator, lang i18n.LanguageCodes) openai.ChatCompletionRequest {
	msgs := deepcopy.Copy(promptStruct.Messages).([]openai.ChatCompletionMessage)

	req := openai.ChatCompletionRequest{
		Messages: msgs,
		Stream:   false,
	}

	var tmplArgs struct {
		CodeLanguage string
		FileName     string
		FileContent  string
		UserLang     string
	}
	tmplArgs.CodeLanguage = r.CodeLanguage
	tmplArgs.FileName = r.FileName
	tmplArgs.FileContent = r.FileContent
	tmplArgs.UserLang = i18nutil.GetUserLang(lang)

	for i := range req.Messages {
		t, _ := template.New("").Parse(req.Messages[i].Content)
		if t == nil {
			continue
		}
		buffer := bytes.NewBuffer(nil)
		_ = t.Execute(buffer, tmplArgs)
		req.Messages[i].Content = buffer.String()
	}

	return req
}
