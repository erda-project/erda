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

package elasticsearch

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
)

type (
	config struct {
		QueryTimeout time.Duration `file:"query_timeout"`
		WriteTimeout time.Duration `file:"write_timeout"`
		IndexType    string        `file:"index_type" default:"default"`
	}
	provider struct {
		Cfg          *config
		Log          logs.Logger
		ES           elasticsearch.Interface `autowired:"elasticsearch"`
		queryTimeout string
		writeTimeout string
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	if p.Cfg.QueryTimeout > 0 {
		p.queryTimeout = fmt.Sprintf("%dms", p.Cfg.QueryTimeout.Milliseconds())
	}
	if p.Cfg.WriteTimeout > 0 {
		p.writeTimeout = fmt.Sprintf("%dms", p.Cfg.WriteTimeout.Milliseconds())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return &esStorage{
		client:       p.ES.Client(),
		typ:          p.Cfg.IndexType,
		queryTimeout: p.queryTimeout,
		writeTimeout: p.writeTimeout,
	}
}

func init() {
	servicehub.Register("elasticsearch-storage", &servicehub.Spec{
		Services:   []string{"elasticsearch-storage-reader", "elasticsearch-storage-writer"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
