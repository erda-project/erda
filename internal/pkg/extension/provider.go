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

package extension

import (
	"context"
	"reflect"
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/extension/db"
)

type config struct {
	ExtensionMenu        map[string][]string `file:"extension_menu" env:"EXTENSION_MENU"`
	ExtensionSources     string              `file:"extension_sources" env:"EXTENSION_SOURCES"`
	ExtensionSourcesCron string              `file:"extension_sources_cron" env:"EXTENSION_SOURCES_CRON"`
	ReloadExtensionType  string              `file:"reload_extension_type" env:"RELOAD_EXTENSION_TYPE"`

	InitFilePath string `file:"init_file_path" default:"common-conf/extensions-init"`
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register `autowired:"service-register" required:"true"`
	DB       *gorm.DB           `autowired:"mysql-client"`

	db            *db.Client
	bdl           *bundle.Bundle
	extensionMenu map[string][]string

	cacheExtensionSpecs sync.Map
}

func (s *provider) Init(ctx servicehub.Context) error {
	s.db = &db.Client{DB: s.DB}
	return nil
}

func (s *provider) Run(ctx context.Context) error {
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("extension", &servicehub.Spec{
		Services:     []string{"extension"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: nil,
		Description:  "public extension service",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
