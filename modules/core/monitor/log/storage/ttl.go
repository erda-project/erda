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
	"strings"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/pkg/router"
)

type ttlStore interface {
	GetSecond(name string, kvs map[string]string) int
	Run(ctx context.Context, interval time.Duration)
}

type mysqlStore struct {
	defaultTTLSec int
	ttlsValue     atomic.Value
	mysql         *gorm.DB
	Log           logs.Logger
}

func (m *mysqlStore) loadLogsTTL() error {
	type MonitorConfig struct {
		OrgName string `gorm:"column:org_name"`
		Names   string `gorm:"column:names"`
		Filters string `gorm:"column:filters"`
		Config  string `gorm:"column:config"`
		Key     string `gorm:"column:key"`
	}
	type monitorConfig struct {
		TTL string `gorm:"column:ttl"`
	}

	var list []*MonitorConfig
	if err := m.mysql.Table("sp_monitor_config").Where("`type`='log' AND `enable`=1").Find(&list).Error; err != nil {
		m.Log.Errorf("fail to load sp_monitor_config: %s", err)
		return err
	}
	r := router.New()
	for _, item := range list {
		if len(item.Names) <= 0 || len(item.Config) <= 0 || len(item.Key) <= 0 {
			continue
		}
		mc := &monitorConfig{}
		err := json.Unmarshal(reflectx.StringToBytes(item.Config), mc)
		if err != nil || len(mc.TTL) <= 0 {
			m.Log.Errorf("invalid monitor log config for key=%s", item.Key)
			continue
		}
		d, err := time.ParseDuration(mc.TTL)
		if err != nil || int64(d) < int64(time.Second) {
			m.Log.Errorf("invalid monitor log config for key=%s, %s", item.Key, err)
			continue
		}
		var filter []*router.KeyValue
		err = json.Unmarshal(reflectx.StringToBytes(item.Filters), &filter)
		if err != nil {
			m.Log.Errorf("invalid monitor log config filters for key=%s", item.Key)
			continue
		}
		for _, name := range strings.Split(item.Names, ",") {
			r.Add(name, filter, int(d.Seconds()))
		}
	}
	m.ttlsValue.Store(r)
	m.Log.Info("load logs ttl config")
	return nil
}

func (m *mysqlStore) GetSecond(name string, kvs map[string]string) int {
	r := m.ttlsValue.Load().(*router.Router)
	if r == nil {
		return m.defaultTTLSec
	}

	val := r.Find(name, kvs)
	if val == nil {
		return m.defaultTTLSec
	}
	return val.(int)
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
