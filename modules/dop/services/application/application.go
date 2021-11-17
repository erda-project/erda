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

package application

import (
	"github.com/erda-project/erda-infra/providers/i18n"
	dashboardPb "github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/bundle"
)

type Application struct {
	bdl *bundle.Bundle
	trans i18n.Translator
	cmp dashboardPb.ClusterResourceServer
}

func New(options ...Option) *Application {
	app := new(Application)
	for _, opt := range options {
		opt(app)
	}
	return app
}

type Option func(application *Application)

func WithBundle(bdl *bundle.Bundle) Option {
	return func(app *Application) {
		app.bdl = bdl
	}
}

func WithTranslator(trans i18n.Translator) Option {
	return func(app *Application) {
		app.trans = trans
	}
}

func WithCMP(cmp dashboardPb.ClusterResourceServer) Option {
	return func(app *Application) {
		app.cmp = cmp
	}
}
