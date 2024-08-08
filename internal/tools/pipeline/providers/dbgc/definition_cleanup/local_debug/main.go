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

package main

import (
	"context"
	"os"
	"time"

	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers/serviceregister"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/cron"
	_ "github.com/erda-project/erda/internal/tools/pipeline/providers/dbgc/definition_cleanup"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/leaderworker"
)

type provider struct {
	LW leaderworker.Interface
}

type config struct {
}

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	time.Sleep(time.Second * 5)
	p.LW.Start()
	return nil
}

func init() {
	servicehub.Register("local-debug", &servicehub.Spec{
		Services:     []string{"local-debug"},
		Dependencies: []string{"definition-cleanup"},
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Description: "local-debug",
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}

func main() {
	hub := servicehub.New()
	hub.Run("local-debug", "", os.Args...)
}
