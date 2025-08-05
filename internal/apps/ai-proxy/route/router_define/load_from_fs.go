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

package router_define

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

type YamlFile struct {
	Routes []*Route `yaml:"routes"`
}

func LoadRoutesFromEmbeddedDir(routesFS embed.FS) (*YamlFile, error) {
	yamlFile := &YamlFile{Routes: []*Route{}}

	err := loadRoutesFromFS(routesFS, yamlFile)
	if err != nil {
		return nil, err
	}

	// Pre-validate whether all filters in routes are registered
	if err := validateFiltersExistence(yamlFile.Routes); err != nil {
		return nil, err
	}

	return yamlFile, nil
}

func loadRoutesFromFS(routesFS fs.FS, yamlFile *YamlFile) error {
	// Walk through all files in the embedded filesystem
	err := fs.WalkDir(routesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process yaml files
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Read file content
		content, err := fs.ReadFile(routesFS, path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", path, err)
		}

		// Parse YAML
		var fileYaml YamlFile
		if err := yaml.Unmarshal(content, &fileYaml); err != nil {
			return fmt.Errorf("failed to parse YAML file %s: %v", path, err)
		}

		// Merge routes
		yamlFile.Routes = append(yamlFile.Routes, fileYaml.Routes...)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk embedded filesystem: %v", err)
	}

	return nil
}

// validateFiltersExistence validates whether all filters configured in routes are registered
func validateFiltersExistence(routes []*Route) error {
	var missingFilters []string
	seenMissing := make(map[string]bool)

	for _, route := range routes {
		// Validate request filters
		for _, filterConfig := range route.RequestFilters {
			if _, exists := filter_define.FilterFactory.RequestFilters[filterConfig.Name]; !exists {
				key := fmt.Sprintf("request:%s", filterConfig.Name)
				if !seenMissing[key] {
					missingFilters = append(missingFilters, fmt.Sprintf("request filter '%s'", filterConfig.Name))
					seenMissing[key] = true
				}
			}
		}

		// Validate response filters
		for _, filterConfig := range route.ResponseFilters {
			if _, exists := filter_define.FilterFactory.ResponseFilters[filterConfig.Name]; !exists {
				key := fmt.Sprintf("response:%s", filterConfig.Name)
				if !seenMissing[key] {
					missingFilters = append(missingFilters, fmt.Sprintf("response filter '%s'", filterConfig.Name))
					seenMissing[key] = true
				}
			}
		}
	}

	if len(missingFilters) > 0 {
		return fmt.Errorf("the following filters are not registered: %s", strings.Join(missingFilters, ", "))
	}

	return nil
}
