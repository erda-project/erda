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
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis"
	v1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

type (
	config struct {
		ClientCacheSize        int           `file:"client_cache_size" default:"128"`
		ClientCacheExpiration  time.Duration `file:"client_cache_expiration" default:"10m"`
		PodInfoCacheSize       int           `file:"pod_info_cache_size" default:"128"`
		PodInfoCacheExpiration time.Duration `file:"pod_info_cache_expiration" default:"3h"`
		BufferLines            int           `file:"buffer_lines" default:"1024"`
		TimeSpan               time.Duration `file:"time_span" default:"3m"`
	}
	provider struct {
		Cfg     *config
		Log     logs.Logger
		Redis   *redis.Client    `autowired:"redis-client"`
		Loader  loader.Interface `autowired:"elasticsearch.index.loader@log"`
		ctx     servicehub.Context
		clients ClientManager
		pods    PodInfoQueryer
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.ctx = ctx
	p.clients = newClientManager(p.Cfg)
	p.pods = newPodInfoQueryer(p)
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return &cStorage{
		log: p.Log,
		getQueryFunc: func(clusterName string) (func(it *logsIterator, opts *v1.PodLogOptions) (io.ReadCloser, error), error) {
			client, err := p.clients.GetClient(clusterName)
			if err != nil {
				return nil, err
			}
			if client == nil {
				return nil, fmt.Errorf("not found clientset")
			}
			return func(it *logsIterator, opts *v1.PodLogOptions) (io.ReadCloser, error) {
				return client.CoreV1().Pods(it.podNamespace).GetLogs(it.podName, opts).Stream(it.ctx)
			}, nil
		},
		pods:        p.pods,
		bufferLines: int64(p.Cfg.BufferLines),
		timeSpan:    int64(p.Cfg.TimeSpan),
	}
}

func init() {
	servicehub.Register("kubernetes-logs-storage", &servicehub.Spec{
		Services:   []string{"log-storage-kubernetes-reader"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
