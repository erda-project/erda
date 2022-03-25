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

package initializer

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"
)

type ddlFile struct {
	Path      string `file:"path"`
	IgnoreErr bool   `file:"ignore_err"`
}

type config struct {
	DDLs     []ddlFile `file:"ddl_files"`
	Database string    `file:"database" default:"monitor"`
}

type provider struct {
	Cfg        *config
	Log        logs.Logger
	Clickhouse clickhouse.Interface `autowired:"clickhouse" inherit-label:"preferred"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	return p.initDDLs()
}

func init() {
	servicehub.Register("clickhouse.table.initializer", &servicehub.Spec{
		Services:     []string{"clickhouse.table.initializer"},
		Dependencies: []string{"clickhouse"},
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
