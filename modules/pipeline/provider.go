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

package pipeline

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	_ "github.com/erda-project/erda-infra/providers/etcd"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	_ "github.com/erda-project/erda/modules/pipeline/aop/plugins"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type provider struct {
	CmsService         pb.CmsServiceServer `autowired:"erda.core.pipeline.cms.CmsService"`
	ReconcilerElection election.Interface  `autowired:"etcd-election@reconciler"`
	GcElection         election.Interface  `autowired:"etcd-election@gc"`
	server             *httpserver.Server
}

func (p *provider) Run(ctx context.Context) error {
	logrus.Infof("[alert] starting pipeline instance")
	var err error
	done := make(chan struct{}, 1)

	go func() {
		err = p.server.ListenAndServe()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}

	return err
}

func init() {
	servicehub.Register("pipeline", &servicehub.Spec{
		Services:     []string{"pipeline"},
		Dependencies: []string{"etcd"},
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
