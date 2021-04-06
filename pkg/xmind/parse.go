// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// package xmind require python3
package xmind

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/erda-project/erda/pkg/filehelper"
)

func Parse(r io.Reader) (Content, error) {
	// 解析 .xmind 文件
	parsedJsonFile, err := getContentJsonReader(r)
	if err != nil {
		return nil, err
	}
	var content Content
	if err := json.NewDecoder(parsedJsonFile).Decode(&content); err != nil {
		return nil, err
	}
	return content, nil
}

// getContentJsonReader 通过 xmindparser 获取 content.json reader
func getContentJsonReader(r io.Reader) (io.Reader, error) {
	// 创建临时目录
	tmpDir := os.TempDir()
	// 创建文件
	baseName := "import.xmind"
	fp := filepath.Join(tmpDir, baseName)
	if err := filehelper.CreateFile2(fp, r, 0755); err != nil {
		return nil, err
	}
	cmd := exec.Command("xmindparser", fp, "-json")
	cmd.Dir = tmpDir
	cmd.Env = append(cmd.Env, `LANG=en_US.UTF-8`)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to use xmindparser parse xmind file, err: %v, output: %s", err, cmdOutput)
	}
	generateJsonFilePath := filepath.Join(tmpDir, "import.json")
	jsonFile, err := os.Open(generateJsonFilePath)
	if err != nil {
		return nil, err
	}
	return jsonFile, nil
}
