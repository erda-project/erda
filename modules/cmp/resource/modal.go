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

package resource

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/modules/cmp/interface"
)

type Resource struct {
	Ctx    context.Context
	Server _interface.Provider `autowired:"erda.cmp"`
	I18N   i18n.Translator
	Lang   i18n.LanguageCodes
	DB     *dbclient.DBClient
}

func (r *Resource) I18n(key string, args ...interface{}) string {
	if len(args) == 0 {
		try := r.I18N.Text(r.Lang, key)
		if try != key {
			return try
		}
	}
	return r.I18N.Sprintf(r.Lang, key, args...)
}

func New(ctx context.Context, i18n i18n.Translator, lang i18n.LanguageCodes) *Resource {
	r := &Resource{}
	r.I18N = i18n
	r.Ctx = ctx
	r.Lang = lang
	return r
}

func (r *Resource) Init(ctx servicehub.Context) error {
	sServer := ctx.Service("cmp").(_interface.Provider)
	r.Server = sServer
	r.Ctx = ctx
	return nil
}

//"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"

type XAixs struct {
	Type string   `json:"type"`
	Data []string `json:"data"`
}

type YAixs struct {
	Type string `json:"type"`
}

type Series struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Data []float64 `json:"data"`
}

type Histogram struct {
	XAixs  XAixs
	YAixs  YAixs
	Series []HistogramSerie `json:"series"`
}
