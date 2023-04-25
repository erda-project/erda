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

package audit_test

import (
	"testing"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter/audit"
)

const (
	createCompletionRespBody = `{
    "id": "cmpl-792ye0BRe3PtyHVOnTwBGlapRzgso",
    "object": "text_completion",
    "created": 1682390752,
    "model": "text-davinci-003",
    "choices": [
        {
            "text": "？\n\nRust 的借用检查器是 Rust 语言的一个核心功能，它通过一系列类型安全的规则来确保在运行时程序的安全性。Rust 专注于编写安全的代码，借用检查器就是实现这个目标的一种方式。Rust 的借用检查器可以检测出在编译期和运行期的可能的错误，避免出现访问释放的内存，使",
            "index": 0,
            "logprobs": null,
            "finish_reason": "length"
        }
    ],
    "usage": {
        "prompt_tokens": 26,
        "completion_tokens": 255,
        "total_tokens": 281
    }
}`
)

func TestExtractCompletionFromCreateCompletionResp(t *testing.T) {
	s, err := audit.ExtractCompletionFromCreateCompletionResp(createCompletionRespBody)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
}
