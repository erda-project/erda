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

package reconciler

import (
	"context"
	"fmt"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/legacy/reconciler"
)

type provider struct {
	Log logs.Logger
	Cfg *config

	MySQL mysqlxorm.Interface
	LW    leaderworker.Interface

	r *reconciler.Reconciler
}

type config struct {
}

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	if p.r == nil {
		return fmt.Errorf("set reconciler before run")
	}

	// gc
	p.LW.OnLeader(p.r.ListenGC)
	p.LW.OnLeader(p.r.CompensateGCNamespaces)

	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("reconciler", &servicehub.Spec{
		Services:     []string{"reconciler"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: nil,
		Description:  "pipeline reconciler",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
