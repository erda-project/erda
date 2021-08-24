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

package storage

import (
	"context"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
)

type ttlStore interface {
	GetSecondByKey(key string) int
	Run(ctx context.Context, interval time.Duration)
}

type mysqlStore struct {
	defaultTTLSec int
	ttlValue      map[string]int
	mysql         *gorm.DB
	Log           logs.Logger
	mu            sync.RWMutex
}

func (m *mysqlStore) GetSecondByKey(key string) int {
	m.mu.RLock()
	ttl, ok := m.ttlValue[key]
	m.mu.RUnlock()
	if ok {
		return ttl
	}
	return m.defaultTTLSec
}

func (m *mysqlStore) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := m.loadLogsTTL(); err != nil {
				m.Log.Errorf("loadLogs failed. err=%s", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

type MonitorConfig struct {
	OrgName string `gorm:"column:org_name"`
	Names   string `gorm:"column:names"`
	Filters string `gorm:"column:filters"`
	Config  []byte `gorm:"column:config"`
	Key     string `gorm:"column:key"`
}

type monitorConfig struct {
	TTL string `gorm:"column:ttl" json:"ttl"`
}

func (m *mysqlStore) loadLogsTTL() error {
	var list []*MonitorConfig
	if err := m.mysql.Table("sp_monitor_config").Where("`type`='log' AND `enable`=1").Find(&list).Error; err != nil {
		m.Log.Errorf("fail to load sp_monitor_config: %s", err)
		return err
	}
	ttlmap := m.populateTTLValue(list)
	m.mu.Lock()
	m.ttlValue = ttlmap
	m.mu.Unlock()
	m.Log.Info("load logs ttl config")
	return nil
}

func (m *mysqlStore) populateTTLValue(list []*MonitorConfig) map[string]int {
	ttlMap := make(map[string]int)
	for _, item := range list {
		var mc monitorConfig
		err := json.Unmarshal(item.Config, &mc)
		if err != nil || len(mc.TTL) <= 0 {
			m.Log.Errorf("invalid monitor log config for key=%s", item.Key)
			continue
		}
		d, err := time.ParseDuration(mc.TTL)
		if err != nil || int64(d) < int64(time.Second) {
			m.Log.Errorf("invalid monitor log config for key=%s, %s", item.Key, err)
			continue
		}
		ttlMap[item.OrgName] = int(d.Seconds())
	}
	return ttlMap
}
