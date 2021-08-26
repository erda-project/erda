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

package conf

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Configure struct {
	FS      embed.FS
	Dirname string
}

const (
	JsonFileExtension = ".json"
	YamlFileExtension = ".yaml"
	YmlFileExtension  = ".yml"
)

func (c *Configure) YamlReader() map[string]*map[string]interface{} {
	files := map[string]*map[string]interface{}{}
	reader(c.FS, c.Dirname, YamlFileExtension, files)
	return files
}

func (c *Configure) JsonReader() map[string]*map[string]interface{} {
	files := map[string]*map[string]interface{}{}
	reader(c.FS, c.Dirname, JsonFileExtension, files)
	return files
}

func reader(fs embed.FS, dirname, fileExtension string, files map[string]*map[string]interface{}) {
	entries, err := fs.ReadDir(dirname)
	if err != nil {
		log.Printf("Read dir(%s) with error: %+v\n", dirname, err)
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Printf("Read fs entry with error: %+v\n", err)
		}

		filename := fmt.Sprintf("%s/%s", dirname, info.Name())
		if info.IsDir() {
			reader(fs, filename, fileExtension, files)
		}

		if !strings.HasSuffix(info.Name(), fileExtension) {
			continue
		}

		file, err := fs.ReadFile(filename)
		if err != nil {
			log.Printf("Read file(%s) with error: %+v\n", filename, err)
			continue
		}
		var expression map[string]interface{}
		err = json.Unmarshal(file, &expression)
		if err != nil {
			log.Printf("Unmarshal file(%s) with error: %+v\n", filename, err)
			continue
		}
		files[expression["id"].(string)] = &expression
	}
}
