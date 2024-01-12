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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
)

func GetFileContent(repo *gitmodule.Repository, mr *apistructs.MergeRequestInfo, filePath string) (fileContent string, truncated bool, err error) {
	ref := mr.SourceBranch + "/" + filePath
	if err = repo.ParseRefAndTreePath(ref); err != nil {
		return
	}
	treeEntry, err := repo.GetParsedTreeEntry()
	if err != nil {
		return
	}
	if treeEntry.IsDir() {
		err = fmt.Errorf("file path is a directory")
		return
	}
	data, err := treeEntry.Blob().Data()
	if err != nil {
		return
	}
	buf := make([]byte, 1024)
	n, _ := data.Read(buf)
	buf = buf[:n]
	contentType := http.DetectContentType(buf)
	isTextFile := isTextType(contentType)
	if !isTextFile {
		err = fmt.Errorf("file is not text file")
		return
	}
	d, err := io.ReadAll(data)
	if err != nil {
		return
	}
	buf = append(buf, d...)
	if len(buf) > MAX_FILE_CHANGES_CHAR_SIZE {
		buf = buf[:MAX_FILE_CHANGES_CHAR_SIZE]
		truncated = true
	}
	fileContent = string(buf)
	return
}

func isTextType(contentType string) bool {
	return strings.Contains(contentType, "text/")
}
