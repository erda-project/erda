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

package uc_adaptor

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type config struct {
	OryEnabled bool `default:"false" file:"ORY_ENABLED" env:"ORY_ENABLED"`
}

type provider struct {
	Cfg *config
}

func (p *provider) Run(ctx context.Context) error {
	// only needed when uc is used.
	if p.Cfg.OryEnabled {
		return nil
	}
	return Initialize()
}

func init() {
	servicehub.Register("uc-adaptor", &servicehub.Spec{
		Services:   []string{"uc-adaptor"},
		Creator:    func() servicehub.Provider { return &provider{} },
		ConfigFunc: func() interface{} { return &config{} },
	})
}
