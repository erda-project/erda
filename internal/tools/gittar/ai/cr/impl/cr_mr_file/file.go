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
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/erda-project/erda-infra/providers/i18n"

	"github.com/mohae/deepcopy"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/impl/cr_mr_code_snippet"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/aiutil"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/mrutil"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/pkg/strutil"
)

type OneChangedFile struct {
	FileName     string
	CodeLanguage string
	Truncated    bool
	CodeSnippets []cr_mr_code_snippet.CodeSnippet

	mr       *apistructs.MergeRequestInfo
	diffFile *gitmodule.DiffFile
	user     *models.User
}

func NewFileReviewer(diffFile *gitmodule.DiffFile, user *models.User, mr *apistructs.MergeRequestInfo) models.FileCodeReviewer {
	fr := OneChangedFile{diffFile: diffFile, user: user, mr: mr}
	fr.FileName = diffFile.Name
	fr.CodeLanguage = strings.TrimPrefix(filepath.Ext(diffFile.Name), ".")
	return &fr
}

func init() {
	models.RegisterCodeReviewer(models.AICodeReviewTypeMRFile, func(req models.AICodeReviewNoteRequest, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, user *models.User) (models.CodeReviewer, error) {
		if req.NoteLocation.NewPath == "" {
			return nil, fmt.Errorf("no file specified")
		}
		diffFile := mrutil.GetDiffFileFromMR(repo, mr, req.NoteLocation.NewPath)
		if diffFile == nil {
			return nil, fmt.Errorf("file not found")
		}
		return NewFileReviewer(diffFile, user, mr), nil
	})
}

func (r *OneChangedFile) GetFileName() string {
	return r.FileName
}

// CodeReview for file level, invoke once with all code snippets.
func (r *OneChangedFile) CodeReview(i18n i18n.Translator, lang i18n.LanguageCodes) string {
	if r.diffFile == nil {
		return ""
	}
	r.parseCodeSnippets()

	// invoke ai
	result := aiutil.InvokeAI(r.constructAIRequest(), r.user)
	if result == "" {
		return ""
	}

	// handle response
	var res FileReviewResult
	if err := json.Unmarshal([]byte(result), &res); err != nil {
		logrus.Warnf("failed to unmarshal ai result, err: %s", err)
	}
	// group result by snippet index
	snippetIndexIssues := make(map[int][]FileReviewResultItem)
	for _, item := range res.Result {
		if item.SnippetIndex >= len(r.CodeSnippets) {
			continue
		}
		snippetIndexIssues[item.SnippetIndex] = append(snippetIndexIssues[item.SnippetIndex], item)
	}
	// handle each snippet index
	var lines []string
	for snippetIndex := range r.CodeSnippets {
		// add original code
		lines = append(lines, fmt.Sprintf("### %s:", i18n.Text(lang, models.I18nKeyCodeSnippet)), r.CodeSnippets[snippetIndex].GetMarkdownCode())
		for _, issue := range snippetIndexIssues[snippetIndex] {
			if ss := strings.Split(issue.Details, "\n"); len(ss) > 1 {
				issue.Details = "\n" + issue.Details
			}
			lines = append(lines, fmt.Sprintf("**%s:** %s", i18n.Text(lang, models.I18nKeyMrAICrIssue), issue.Details), "")
			lines = append(lines, fmt.Sprintf("**%s:** %s", i18n.Text(lang, models.I18nKeyMrAICrRiskLevel), i18n.Text(lang, models.I18nKeyMrAICrRiskLevelPrefix+issue.RiskLevel)), "")
		}
		lines = append(lines, "---", "")
	}
	return strings.Join(lines, "\n")
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

//go:embed fc.yaml
var functionDefinitionYaml string
var functionDefinition json.RawMessage

type PromptStruct = struct {
	Messages []openai.ChatCompletionMessage `yaml:"messages"`
}

var promptStruct PromptStruct

func init() {
	if err := yaml.Unmarshal([]byte(promptYaml), &promptStruct); err != nil {
		panic(err)
	}
	var err error
	functionDefinition, err = strutil.YamlOrJsonToJson([]byte(functionDefinitionYaml))
	if err != nil {
		panic(err)
	}
}

func (r *OneChangedFile) constructAIRequest() openai.ChatCompletionRequest {
	msgs := deepcopy.Copy(promptStruct.Messages).([]openai.ChatCompletionMessage)

	req := openai.ChatCompletionRequest{
		Messages: msgs,
		Stream:   false,
		Functions: []openai.FunctionDefinition{
			{
				Name:        "create-cr-note",
				Description: "create code review note",
				Parameters:  functionDefinition,
			},
		},
		FunctionCall: openai.FunctionCall{
			Name: "create-cr-note",
		},
	}

	var tmplArgs struct {
		CodeLanguage string
		FileName     string
		FileContents string
	}
	tmplArgs.CodeLanguage = r.CodeLanguage
	tmplArgs.FileName = r.FileName

	type SnippetContent struct {
		SnippetIndex  int    `json:"snippetIndex"`
		PromptContent string `json:"promptContent"`
	}
	var changedFileContents []string
	for i, cs := range r.CodeSnippets {
		sc := SnippetContent{SnippetIndex: i, PromptContent: cs.SelectedCode}
		b, _ := json.Marshal(sc)
		changedFileContents = append(changedFileContents, string(b))
	}
	tmplArgs.FileContents = strings.Join(changedFileContents, "\n\n")

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

func (r *OneChangedFile) parseCodeSnippets() {
	for _, section := range r.diffFile.Sections {
		selectedCode, truncated := mrutil.ConvertDiffLinesToSnippet(section.Lines)
		if strings.TrimSpace(selectedCode) == "" {
			continue
		}
		codeSnippet := cr_mr_code_snippet.CodeSnippet{
			CodeLanguage: strings.TrimPrefix(filepath.Ext(r.diffFile.Name), "."),
			SelectedCode: selectedCode,
			Truncated:    truncated,
		}
		r.CodeSnippets = append(r.CodeSnippets, codeSnippet)
	}
}
