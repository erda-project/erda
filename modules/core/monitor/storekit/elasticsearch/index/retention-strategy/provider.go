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

package retention

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/cleaner"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

// Interface .
type Interface interface {
	GetTTL(key string) time.Duration
	DefaultTTL() time.Duration
	GetConfigKey(name string, tags map[string]string) string
	Loading(ctx context.Context)
}

type (
	config struct {
		Pattern string `file:"pattern"`
	}
	provider struct {
		Cfg      *config
		Log      logs.Logger
		strategy Interface `autowired:"storage-retention-strategy"`
		pattern  *index.Pattern
		keyIdx   int
	}
)

var _ cleaner.RetentionStrategy = (*provider)(nil)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	ptn, err := index.BuildPattern(p.Cfg.Pattern)
	if err != nil {
		return err
	}
	p.keyIdx = -1
	for i, key := range ptn.Keys {
		if key == "key" {
			p.keyIdx = i
			break
		}
	}
	if p.keyIdx < 0 {
		return fmt.Errorf("not fount <key> in pattern %q", p.Cfg.Pattern)
	}
	p.pattern = ptn
	return p.initStrategy(ctx)
}

func (p *provider) GetTTL(entry *loader.IndexEntry) time.Duration {
	result, ok := p.pattern.Match(entry.Index, index.InvalidPatternValueChars)
	if ok {
		return p.strategy.GetTTL(result.Keys[p.keyIdx])
	}
	return p.strategy.DefaultTTL()
}

func (p *provider) Loading(ctx context.Context) {
	p.strategy.Loading(ctx)
}

func (p *provider) initStrategy(ctx servicehub.Context) error {
	obj, name := index.FindService(ctx, "storage-retention-strategy")
	if obj == nil {
		return fmt.Errorf("%q is required", name)
	}
	strategy, ok := obj.(Interface)
	if !ok {
		return fmt.Errorf("%q is not StorageRetentionStrategy", name)
	}
	p.strategy = strategy
	p.Log.Debugf("use StorageRetentionStrategy (%q) for index clean", name)
	return nil
}

func init() {
	servicehub.Register("elasticsearch.index.retention-strategy", &servicehub.Spec{
		Services:   []string{"elasticsearch.index.retention-strategy"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
