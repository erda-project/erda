// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"context"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/jinzhu/gorm"
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
