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

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
)

type Config struct {
	LogLevelStr string       `file:"log_level" default:"info" env:"LOG_LEVEL"`
	LogLevel    logrus.Level `json:"-" yaml:"-"`

	SelfURL    string `file:"self_url" env:"SELF_URL" required:"true"`
	OpenOnErda bool   `file:"open_on_erda" default:"false" env:"OPEN_ON_ERDA"`

	Exporter configPromExporter

	RoutesRefs  []string     `file:"routes_refs"`
	Routes      route.Routes `json:"-" yaml:"-"`
	MaxFileSize uint64       `file:"max_file_size"`
}

type configPromExporter struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Subsystem string `json:"subsystem" yaml:"subsystem"`
	Name      string `json:"name" yaml:"name"`
}

// DoPost do some post process after config loaded
func (cfg *Config) DoPost() error {
	// parse routes refs
	for _, routeRef := range cfg.RoutesRefs {
		var routes route.Routes
		if err := parseFileConfig(routeRef, "routes", &routes); err != nil {
			return fmt.Errorf("failed to parse routes ref: %v", err)
		}
		cfg.Routes = append(cfg.Routes, routes...)
	}
	if err := cfg.Routes.Validate(); err != nil {
		return fmt.Errorf("failed to validate routes: %v", err)
	}

	// parse log level
	level, err := logrus.ParseLevel(cfg.LogLevelStr)
	if err != nil {
		return fmt.Errorf("failed to parse log level, level: %s, err: %v", cfg.LogLevel, err)
	}
	logrus.SetLevel(level)
	cfg.LogLevel = level

	return nil
}

func parseFileConfig(refPath, key string, target interface{}) error {
	data, err := os.ReadFile(refPath)
	if err != nil {
		return err
	}
	var m = make(map[string]json.RawMessage)
	if err := yaml.Unmarshal(data, &m); err != nil {
		return err
	}
	data, ok := m[key]
	if !ok {
		return nil
	}
	return yaml.Unmarshal(data, target)
}
