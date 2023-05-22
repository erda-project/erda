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
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda/pkg/router"
)

func (p *provider) getConfig(ctx context.Context) *retentionConfig {
	v, _ := p.value.Load().(*retentionConfig)
	for v == nil {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(1 * time.Second):
		}
		v, _ = p.value.Load().(*retentionConfig)
	}
	return v
}

// GetTTL return hot and all ttl
func (p *provider) GetTTL(key string) *TTL {
	cfg := p.getConfig(context.Background())
	d, ok := cfg.keysTTL[key]
	if ok {
		if d.HotData <= time.Duration(0) {
			d.HotData = p.Cfg.DefaultHotTTL
		}
		return d
	}
	return &TTL{
		HotData: p.Cfg.DefaultHotTTL,
		All:     p.Cfg.DefaultTTL,
	}
}

func (p *provider) Default() *TTL {
	return &TTL{
		HotData: p.Cfg.DefaultHotTTL,
		All:     p.Cfg.DefaultTTL,
	}
}
func (p *provider) DefaultTTL() time.Duration {
	return p.Cfg.DefaultTTL
}

func (p *provider) DefaultHotDataTTL() time.Duration {
	return p.Cfg.DefaultHotTTL
}

func (p *provider) GetConfigKey(name string, tags map[string]string) string {
	cfg := p.getConfig(context.Background())
	matched := cfg.matcher.Find(name, tags)
	if matched != nil {
		c := matched.(*configItem)
		return c.Key
	}
	return ""
}

func (p *provider) GetTTLByTags(name string, tags map[string]string) time.Duration {
	cfg := p.getConfig(context.Background())
	matched := cfg.matcher.Find(name, tags)
	if matched != nil {
		c := matched.(*configItem)
		return c.Duration.All //old code, compute day index by elasticsearch
	}
	return p.Cfg.DefaultTTL
}

type (
	configItem struct {
		Duration *TTL
		Key      string
	}
	retentionConfig struct {
		matcher *router.Router
		keysTTL map[string]*TTL
	}
)

type configData struct {
	TTL    string `json:"ttl"`
	HotTTL string `json:"hot_ttl"`
}

func (d *configData) unmarshal(configString string) error {
	return json.Unmarshal(reflectx.StringToBytes(configString), d)
}

func (d *configData) getTTL() (*TTL, error) {
	if len(d.TTL) <= 0 {
		return nil, errors.New("ttl should by more than the zero")
	}
	ttl := &TTL{}
	var err error

	var dur = time.Duration(0)
	if d.HotTTL != "" {
		if dur, err = getDuration(d.HotTTL); err != nil {
			return nil, errors.Wrap(err, "hot ttl should by in duration")
		}
	}
	ttl.HotData = dur

	if dur, err = getDuration(d.TTL); err != nil {
		return nil, errors.Wrap(err, "ttl should by in duration")
	}
	ttl.All = dur
	return ttl, nil
}

func getDuration(duration string) (time.Duration, error) {
	durs, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	if int64(durs) <= int64(time.Minute) {
		return 0, errors.Errorf("too small %s retention ttl", durs)
	}
	return durs, nil
}

func (p *provider) loadConfig() error {
	type MonitorConfig struct {
		Names   string `gorm:"column:names"`
		Filters string `gorm:"column:filters"`
		Config  string `gorm:"column:config"`
		Key     string `gorm:"column:key"`
	}
	var list []*MonitorConfig
	if err := p.DB.Table("sp_monitor_config").Where("`type`=? AND `enable`=1", p.typ).Find(&list).Error; err != nil {
		p.Log.Errorf("failed to load sp_monitor_config: %s", err)
		return err
	}

	rc := &retentionConfig{
		matcher: router.New(),
		keysTTL: make(map[string]*TTL),
	}
	for _, item := range list {
		if len(item.Key) <= 0 || len(item.Config) <= 0 {
			continue
		}

		cfg := &configData{}
		err := cfg.unmarshal(item.Config)

		if err != nil {
			p.Log.Errorf("invalid %s retention config, key=%s", p.typ, item.Key)
			continue
		}
		var ttl *TTL
		if ttl, err = cfg.getTTL(); err != nil {
			p.Log.Warnf("invalid %s retention ttl, key=%s: %s", p.typ, item.Key, err)
			continue
		}

		rc.keysTTL[item.Key] = ttl

		var filter []*router.KeyValue
		err = json.Unmarshal(reflectx.StringToBytes(item.Filters), &filter)
		if err != nil {
			p.Log.Errorf("invalid %s retention filters, key=%s", p.typ, item.Key)
			continue
		}
		for _, name := range strings.Split(item.Names, ",") {
			rc.matcher.Add(name, filter, &configItem{
				Duration: ttl,
				Key:      item.Key,
			})
		}
	}
	p.value.Store(rc)
	p.Log.Infof("load %s retention config ok", p.typ)

	if p.Cfg.PrintDetails {
		rc.matcher.PrintTree(false)
	}
	return nil
}
