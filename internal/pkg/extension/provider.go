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
	"net/http"
	"reflect"
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/http/encoding"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda-proto-go/core/extension/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/extension/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/i18n"
)

type config struct {
	ExtensionMenu        map[string][]string `file:"extension_menu" env:"EXTENSION_MENU"`
	ExtensionSources     string              `file:"extension_sources" env:"EXTENSION_SOURCES"`
	ExtensionSourcesCron string              `file:"extension_sources_cron" env:"EXTENSION_SOURCES_CRON"`
	ReloadExtensionType  string              `file:"reload_extension_type" env:"RELOAD_EXTENSION_TYPE"`
	InitFilePath         string              `file:"init_file_path" default:"common-conf/extensions-init"`
	EnableService        bool                `file:"enable_service" default:"false" env:"ERDA_CORE_EXTENSION_ENABLE_SERVICE"`
}

// +provider
type provider struct {
	Cfg                   *config
	Log                   logs.Logger
	Register              transport.Register `autowired:"service-register" required:"true"`
	DB                    *gorm.DB           `autowired:"mysql-client"`
	InitExtensionElection election.Interface `autowired:"etcd-election@initExtension"`

	db            *db.Client
	bdl           *bundle.Bundle
	extensionMenu map[string][]string

	cacheExtensionSpecs sync.Map
}

func (s *provider) Init(ctx servicehub.Context) error {
	s.bdl = bundle.New(bundle.WithErdaServer())
	s.db = &db.Client{DB: s.DB}
	if s.Register != nil && s.Cfg.EnableService {
		pb.RegisterExtensionServiceImp(s.Register, s, apis.Options(),
			transport.WithHTTPOptions(
				transhttp.WithDecoder(func(r *http.Request, data interface{}) error {
					lang := r.URL.Query().Get("lang")
					if lang != "" {
						r.Header.Set("lang", lang)
					}
					locale := i18n.GetLocaleNameByRequest(r)
					if locale != "" {
						i18n.SetGoroutineBindLang(locale)
					} else {
						i18n.SetGoroutineBindLang(i18n.ZH)
					}
					return encoding.DecodeRequest(r, data)
				}),
			))
	}
	return nil
}

func (s *provider) Run(ctx context.Context) error {
	var once sync.Once
	s.InitExtensionElection.OnLeader(func(ctx context.Context) {
		once.Do(func() {
			err := s.InitSources()
			if err != nil {
				panic(err)
			}
		})
	})
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("erda.core.extension", &servicehub.Spec{
		Services:     []string{"extension"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: nil,
		Description:  "extension service",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
