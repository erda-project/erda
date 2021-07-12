// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package synccache

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda/modules/msp/apm/checker/storage/cache"
	"github.com/erda-project/erda/modules/msp/apm/checker/storage/db"
)

type config struct {
	CacheKey     string        `file:"cache_key" default:"checkers"`
	DelayOnStart time.Duration `file:"delay_on_start" default:"1m"`
	Interval     time.Duration `file:"interval" default:"5m"`
}

// +provider
type provider struct {
	Cfg       *config
	Log       logs.Logger
	Redis     *redis.Client      `autowired:"redis-client"`
	DB        *gorm.DB           `autowired:"mysql-client"`
	Election  election.Interface `autowired:"etcd-election"`
	checkerDB *db.CheckerDB
	cache     *cache.Cache
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.checkerDB = &db.CheckerDB{DB: p.DB, ScopeInfoUpdateInterval: 3 * p.Cfg.Interval}
	p.cache = cache.New(p.Cfg.CacheKey, p.Redis)
	p.Election.OnLeader(p.runSycn)
	return nil
}

func (p *provider) runSycn(ctx context.Context) {
	select {
	case <-time.After(p.Cfg.DelayOnStart):
	case <-ctx.Done():
		return
	}
	err := p.doSync()
	if err != nil {
		p.Log.Errorf("fail to sync: %s", err)
	}

	for {
		select {
		case <-time.After(p.Cfg.Interval):
			err := p.doSync()
			if err != nil {
				p.Log.Errorf("fail to sync: %s", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *provider) doSync() error {
	start := time.Now()
	checkers, deleted, err := p.checkerDB.FullList()
	if err != nil {
		return err
	}
	for _, id := range deleted {
		err := p.cache.Remove(id)
		if err != nil {
			p.Log.Errorf("fail to remove key(%d) from cache: %s", id, err)
		}
	}
	for _, c := range checkers {
		err := p.cache.Put(c)
		if err != nil {
			p.Log.Errorf("fail to put key(%d) into cache: %s", c.Id, err)
		}
	}
	p.Log.Infof("checkers sync finished, duration: %s", time.Now().Sub(start))
	return nil
}

func init() {
	servicehub.Register("erda.msp.apm.checker.storage.cache.sync", &servicehub.Spec{
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
