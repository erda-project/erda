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
	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/aiutil"
	"github.com/erda-project/erda/internal/tools/gittar/models"
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

func newSnippetCodeReviewer(codeLang, selectedCode string, truncated bool, user *models.User) CodeReviewer {
	return CodeSnippet{
		CodeLanguage: codeLang,
		SelectedCode: selectedCode,
		Truncated:    truncated,
		user:         user,
	}
}

func (cs CodeSnippet) CodeReview() string {
	// invoke ai
	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: snippetReviewTemplate + cs.getMarkdownCode(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Reply in Chinese.",
			},
		},
		Stream: false,
	}

	// invoke
	return aiutil.InvokeAI(req, cs.user)
}

var snippetReviewTemplate = "Please give a review suggestions for the selected code, use markdown title for each suggestion. Code examples can be provided when necessary." +
	"The first-level title is: AI Code Review.\n" +
	"Code:\n"

func (cs CodeSnippet) getMarkdownCode() string {
	return "```" + cs.CodeLanguage + "\n" + cs.SelectedCode + "\n```"
}
