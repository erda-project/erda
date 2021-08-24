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

package indexmanager

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda/modules/core/monitor/metric"
	"github.com/erda-project/erda/pkg/router"
)

type metricConfig struct {
	matcher *router.Router
	keysTTL map[string]time.Duration
}

func (m *IndexManager) loadConfig() error {
	// MonitorConfig .
	type MonitorConfig struct {
		Names   string `gorm:"column:names"`
		Filters string `gorm:"column:filters"`
		Config  string `gorm:"column:config"`
		Key     string `gorm:"column:key"`
	}
	type configData struct {
		TTL string `json:"ttl"`
	}
	var list []*MonitorConfig
	if err := m.db.Table("sp_monitor_config").Where("`type`='metric' AND `enable`=1").Find(&list).Error; err != nil {
		m.log.Errorf("fail to load sp_monitor_config: %s", err)
		return err
	}
	mc := &metricConfig{
		matcher: router.New(),
		keysTTL: make(map[string]time.Duration),
	}
	for _, item := range list {
		if len(item.Key) <= 0 || len(item.Config) <= 0 {
			continue
		}

		cd := &configData{}
		err := json.Unmarshal(reflectx.StringToBytes(item.Config), cd)
		if err != nil || len(cd.TTL) <= 0 {
			m.log.Errorf("invalid monitor metric config for key=%s: %s", item.Key, err)
			continue
		}
		d, err := time.ParseDuration(cd.TTL)
		if err != nil {
			m.log.Errorf("invalid monitor metric config for key=%s: %s", item.Key, err)
			continue
		}
		if int64(d) <= int64(time.Minute) {
			m.log.Errorf("too small ttl monitor metric config for key=%s, %s", item.Key, cd.TTL)
			continue
		}
		if int64(mc.keysTTL[item.Key]) < int64(d) {
			mc.keysTTL[item.Key] = d
		}

		if len(item.Names) <= 0 {
			continue
		}
		var filters []*router.KeyValue
		err = json.Unmarshal(reflectx.StringToBytes(item.Filters), &filters)
		if err != nil {
			m.log.Errorf("invalid monitor metric config filters for key=%s", item.Key)
			continue
		}
		for _, name := range strings.Split(item.Names, ",") {
			mc.matcher.Add(name, filters, item.Key)
		}
	}
	m.iconfig.Store(mc)
	// mc.matcher.PrintTree(false)
	m.log.Infof("load metrics ttl config with keys: %d", len(mc.keysTTL))
	return nil
}

func (m *IndexManager) getMetricConfig() *metricConfig {
	if m.cfg.LoadIndexTTLFromDatabase {
		var v interface{}
		for {
			v = m.iconfig.Load()
			if v == nil {
				// Waiting for the load to complete
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}
		return v.(*metricConfig)
	}
	return nil
}

func (m *IndexManager) getKey(metric *metric.Metric) string {
	mc := m.getMetricConfig()
	if mc != nil {
		key := mc.matcher.Find(metric.Name, metric.Tags)
		if key != nil {
			return key.(string)
		}
	}
	return ""
}
