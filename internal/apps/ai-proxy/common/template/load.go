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

package template

import (
	"embed"
	"encoding/json"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/template/templatetypes"
)

// LoadTemplatesFromEmbeddedFS reads template definitions from the embedded filesystem.
func LoadTemplatesFromEmbeddedFS(logger logs.Logger, templatesFS embed.FS) (templatetypes.TemplatesByType, error) {
	return LoadTemplatesFromFS(logger, templatesFS)
}

// LoadTemplatesFromFS reads template definitions from any filesystem.
func LoadTemplatesFromFS(logger logs.Logger, templatesFS iofs.FS) (templatetypes.TemplatesByType, error) {
	templatesByType := templatetypes.TemplatesByType{
		templatetypes.TemplateTypeServiceProvider: make(templatetypes.TemplateSet),
		templatetypes.TemplateTypeModel:           make(templatetypes.TemplateSet),
	}

	err := iofs.WalkDir(templatesFS, ".", func(path string, d iofs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		if ext := strings.ToLower(filepath.Ext(path)); ext != ".json" {
			return nil
		}

		content, err := iofs.ReadFile(templatesFS, path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}

		var rawTemplates map[string]json.RawMessage
		if err := json.Unmarshal(content, &rawTemplates); err != nil {
			return fmt.Errorf("failed to parse template file %s: %w", path, err)
		}

		tplType, err := templatetypes.DetectTemplateType(path)
		if err != nil {
			return err
		}
		if _, ok := templatesByType[tplType]; !ok {
			templatesByType[tplType] = make(templatetypes.TemplateSet)
		}

		for templateName, raw := range rawTemplates {
			if _, ok := templatesByType[tplType][templateName]; ok {
				return fmt.Errorf("duplicate template %q detected in type %q", templateName, tplType)
			}

			template := &pb.Template{}
			if err := protojson.Unmarshal(raw, template); err != nil {
				return fmt.Errorf("failed to decode template %q in file %s: %w", templateName, path, err)
			}

			templatesByType[tplType][templateName] = template

			if logger != nil {
				logger.Infof("load template: type: %s, name: %s", tplType, templateName)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk templates filesystem: %w", err)
	}

	return templatesByType, nil
}
