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

package assets

import (
	"fmt"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"os"
	"path/filepath"
	"strings"
)

// Upload return public download url for LLM use.
func Upload(fileName string, data []byte) (string, error) {
	fileUUID, fileStorePath := genFileStorePath(fileName)
	l.Lock()
	assetsMap[fileUUID] = fileStorePath
	l.Unlock()
	// store file
	targetFile, err := os.OpenFile(fileStorePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer targetFile.Close()
	if _, err := targetFile.Write(data); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	return os.Getenv("SELF_PUBLIC_URL") + "/api/ai-proxy/assets/" + fileUUID, nil
}

func genFileStorePath(fileName string) (string, string) {
	fileUUID := uuid.New()
	return fileUUID, filepath.Join(AssetFileDir, fileUUID+"___"+fileName)
}

func getFileDisplayName(fileStorePath string) string {
	ss := strings.SplitN(fileStorePath, "___", 2)
	if len(ss) != 2 {
		return fileStorePath
	}
	return ss[1]
}
