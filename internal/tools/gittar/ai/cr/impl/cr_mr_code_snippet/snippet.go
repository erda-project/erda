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

package cr_mr_code_snippet

import (
	"bytes"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mohae/deepcopy"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/aiutil"
	"github.com/erda-project/erda/internal/tools/gittar/models"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

// CodeSnippet
// 最小 review 单元为 CodeSnippet
// 1. 当用户选择了一段代码，这本身就是 CodeSnippet
// 2. 用户选择了一个文件，那么会有多个 CodeSnippet
// 3. 用户选择整个 mrReviewer，那么需要有多个 CodeSnippet
type CodeSnippet struct {
	CodeLanguage string
	SelectedCode string
	Truncated    bool // if there are too many changes, we have to truncate the content according to the model context

	user *models.User
}

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
	// register
	models.RegisterCodeReviewer(models.AICodeReviewTypeMRCodeSnippet, func(req models.AICodeReviewNoteRequest, repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, user *models.User) (models.CodeReviewer, error) {
		if req.CodeSnippetRelated == nil || req.CodeSnippetRelated.SelectedCode == "" {
			return nil, fmt.Errorf("no code selected")
		}
		if req.CodeSnippetRelated.CodeLanguage == "" && req.NoteLocation.NewPath != "" {
			ext := strings.TrimPrefix(filepath.Ext(req.NoteLocation.NewPath), ".")
			req.CodeSnippetRelated.CodeLanguage = ext
		}
		return newSnippetCodeReviewer(req.CodeSnippetRelated.CodeLanguage, req.CodeSnippetRelated.SelectedCode, false, user), nil
	})
}

func newSnippetCodeReviewer(codeLang, selectedCode string, truncated bool, user *models.User) models.CodeReviewer {
	cs := CodeSnippet{
		CodeLanguage: codeLang,
		SelectedCode: selectedCode,
		Truncated:    truncated,
		user:         user,
	}
	cs.SelectedCode = cs.GetMarkdownCode()
	return cs
}

func (cs CodeSnippet) CodeReview() string {
	// invoke ai
	req := cs.constructAIRequest()

	// invoke
	return aiutil.InvokeAI(req, cs.user)
}

func (cs CodeSnippet) constructAIRequest() openai.ChatCompletionRequest {
	msgs := deepcopy.Copy(promptStruct.Messages).([]openai.ChatCompletionMessage)

	// invoke ai
	req := openai.ChatCompletionRequest{
		Messages: msgs,
		Stream:   false,
	}

	for i := range req.Messages {
		t, _ := template.New("").Parse(req.Messages[i].Content)
		if t == nil {
			continue
		}
		buffer := bytes.NewBuffer(nil)
		_ = t.Execute(buffer, cs)
		req.Messages[i].Content = buffer.String()
	}

	return req
}

func (cs CodeSnippet) GetMarkdownCode() string {
	return "```" + cs.CodeLanguage + "\n" + cs.SelectedCode + "\n```"
}
