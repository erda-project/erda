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

package templatetypes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
)

// TemplateType represents the category a template belongs to. The type
// is usually derived from the directory name under cmd/ai-proxy/conf/templates
// (e.g. "service_provider", "model").
type TemplateType string

// Predefined template types.
const (
	TemplateTypeServiceProvider TemplateType = "service_provider"
	TemplateTypeModel           TemplateType = "model"
)

func IsValidTemplateType(typ string) bool {
	switch TemplateType(typ) {
	case TemplateTypeServiceProvider, TemplateTypeModel:
		return true
	}
	return false
}

type TemplatesByType map[TemplateType]map[string]*pb.Template

// TemplateSet keeps templates keyed by their template name.
type TemplateSet map[string]*pb.Template

func DetectTemplateType(path string) (TemplateType, error) {
	cleanPath := strings.TrimPrefix(filepath.ToSlash(path), "./")
	parts := strings.Split(cleanPath, "/")

	for _, part := range parts {
		switch part {
		case string(TemplateTypeServiceProvider):
			return TemplateTypeServiceProvider, nil
		case string(TemplateTypeModel):
			return TemplateTypeModel, nil
		}
	}

	return "", fmt.Errorf("unknown template type: %s", cleanPath)
}
