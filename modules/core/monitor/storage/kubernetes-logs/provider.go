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

package kuberneteslogs

import (
	"context"
	"io"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/pkg/k8sclient"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type (
	config struct {
		ClusterName string `file:"cluster_name"`
	}
	provider struct {
		Cfg    *config
		Log    logs.Logger
		client *kubernetes.Clientset
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	client, err := k8sclient.New(p.Cfg.ClusterName)
	if err != nil {
		return nil
	}
	p.client = client.ClientSet
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return &cStorage{
		queryFunc: func(ctx context.Context, namespace, pod string, opts *v1.PodLogOptions) (io.ReadCloser, error) {
			return p.client.CoreV1().Pods(namespace).GetLogs(pod, opts).Stream(ctx)
		},
		bufferLines: defaultBufferLines,
	}
}

func init() {
	servicehub.Register("kubernetes-logs-storage", &servicehub.Spec{
		Services:   []string{"kubernetes-logs-storage-reader"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
