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
	"testing"

	"github.com/erda-project/erda/internal/tools/gittar/models"
)

func TestCodeSnippet_GetMarkdownCode(t *testing.T) {
	type fields struct {
		CodeLanguage string
		SelectedCode string
		Truncated    bool
		User         *models.User
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "no lang",
			fields: fields{
				CodeLanguage: "",
				SelectedCode: `fmt.Println("hello world")`,
			},
			want: "```\nfmt.Println(\"hello world\")\n```",
		},
		{
			name: "have lang",
			fields: fields{
				CodeLanguage: "go",
				SelectedCode: `fmt.Println("hello world")`,
			},
			want: "```go\nfmt.Println(\"hello world\")\n```",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := CodeSnippet{
				CodeLanguage: tt.fields.CodeLanguage,
				SelectedCode: tt.fields.SelectedCode,
				Truncated:    tt.fields.Truncated,
				user:         tt.fields.User,
			}
			if got := cs.GetMarkdownCode(); got != tt.want {
				t.Errorf("GetMarkdownCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
