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
	"bufio"
	"bytes"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

const diffLineExamples = `
oldLineNo: -1, newLineNo: -1, type: "section"
oldLineNo:  1, newLineNo: 1,  type: "context"
oldLineNo:  2, newLineNo: 2,  type: "context"
oldLineNo: -1, newLineNo: 3,  type: "add"
oldLineNo: -1, newLineNo: 4,  type: "add"
oldLineNo: -1, newLineNo: 5,  type: "add"
oldLineNo: -1, newLineNo: 6,  type: "add"
oldLineNo: -1, newLineNo: 7,  type: "add"
oldLineNo: -1, newLineNo: 8,  type: "add"
oldLineNo:  3, newLineNo: 9,  type: "context"
oldLineNo:  4, newLineNo: 10, type: "context"
oldLineNo:  5, newLineNo: 11, type: "context"
oldLineNo:  6, newLineNo: -1, type: "delete"
oldLineNo:  7, newLineNo: -1, type: "delete"
oldLineNo: -1, newLineNo: 12, type: "add"
`

func Test_findDiffSectionInOneFile(t *testing.T) {
	// convert diffLineExamples to []*gitmodule.DiffLine
	scanner := bufio.NewScanner(bytes.NewReader([]byte(diffLineExamples)))
	var diffLines []*gitmodule.DiffLine
	var validLineCount int
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		validLineCount++
		items := strings.Split(line, ",")
		var oldLineNo, newLineNo int
		var typ string
		for _, item := range items {
			item = strings.TrimSpace(item)
			kv := strings.Split(item, ":")
			k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			switch k {
			case "oldLineNo":
				_oldLineNo, _ := strconv.ParseInt(v, 10, 64)
				oldLineNo = int(_oldLineNo)
			case "newLineNo":
				_newLineNo, _ := strconv.ParseInt(v, 10, 64)
				newLineNo = int(_newLineNo)
			case "type":
				typ = v
			}
		}
		diffLine := gitmodule.DiffLine{
			OldLineNo: oldLineNo,
			NewLineNo: newLineNo,
			Type:      gitmodule.DiffLineType(typ),
		}
		diffLines = append(diffLines, &diffLine)
	}

	diffFile := &gitmodule.DiffFile{
		Sections: []*gitmodule.DiffSection{
			{
				Lines: diffLines,
			},
		},
	}

	type args struct {
		req NoteRequest
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		wantDiffLineLen int
	}{
		{
			name:            "invalid: oldLineTo and newLineTo must be 0 at the same time",
			args:            args{req: NoteRequest{OldLineTo: 0, NewLineTo: 1}},
			wantErr:         true,
			wantDiffLineLen: 0,
		},
		{
			name:            "invalid: oldLineTo < oldLine, when oldLineTo != -1",
			args:            args{req: NoteRequest{OldLine: 4, OldLineTo: 3, NewLine: 4, NewLineTo: -1}},
			wantErr:         true,
			wantDiffLineLen: 0,
		},
		{
			name:            "valid: oldLine == newLine == -1, means section",
			args:            args{req: NoteRequest{OldLine: -1, NewLine: -1}},
			wantErr:         false,
			wantDiffLineLen: 1,
		},
		{
			name:            "valid: oldLine=1, newLineNo=1, oldLineTo=-1, newLineTo=12",
			args:            args{req: NoteRequest{OldLine: 1, NewLine: 1, OldLineTo: -1, NewLineTo: 12}},
			wantErr:         false,
			wantDiffLineLen: validLineCount,
		},
		{
			name:            "valid: oldLine=-1, newLineNo=4, oldLineTo=4, newLineTo=10",
			args:            args{req: NoteRequest{OldLine: -1, NewLine: 4, OldLineTo: 4, NewLineTo: 10}},
			wantErr:         false,
			wantDiffLineLen: 7 + RelatedDiffLinesCountBefore,
		},
		{
			name:            "valid: oneline, oldLine=3, newLineNo=9, oldLineTo=3, newLineTo=9",
			args:            args{req: NoteRequest{OldLine: 3, NewLine: 9, OldLineTo: 3, NewLineTo: 9}},
			wantErr:         false,
			wantDiffLineLen: 1 + RelatedDiffLinesCountBefore,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findSection, findDiffLines, err := findDiffSectionInOneFile(tt.args.req, diffFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("findDiffSectionInOneFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if len(findDiffLines) != tt.wantDiffLineLen {
					t.Errorf("findDiffSectionInOneFile() gotDiffLineLen = %v, want %v", len(findDiffLines), tt.wantDiffLineLen)
				}
				if !reflect.DeepEqual(findSection, diffFile.Sections[0]) {
					t.Errorf("findDiffSectionInOneFile() gotSection = %v, want %v", findSection, diffFile.Sections[0])
				}
			}
		})
	}
}
