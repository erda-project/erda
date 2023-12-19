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

package mrutil

import (
	"strings"

	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

const MAX_FILE_CHANGES_CHAR_SIZE = 9 * 1000

func ConvertDiffLinesToSnippet(diffLines []*gitmodule.DiffLine) (selectedCode string, truncated bool) {
	var changes []string
	truncated = false
	for _, line := range diffLines {
		s := line.Content
		switch line.Type {
		case gitmodule.DIFF_LINE_ADD:
			s = "+" + s
		case gitmodule.DIFF_LINE_DEL:
			s = "-" + s
		case "": // ignore: \ No newline at end of file
			continue
		case gitmodule.DIFF_LINE_SECTION:
		default:
			s = " " + s
		}
		if s == "" {
			continue
		}
		changes = append(changes, s)
	}
	if len(changes) == 0 {
		return
	}
	if len(changes) > MAX_FILE_CHANGES_CHAR_SIZE {
		changes = changes[:MAX_FILE_CHANGES_CHAR_SIZE]
		truncated = true
	}
	selectedCode = strings.Join(changes, "\n")
	return
}
