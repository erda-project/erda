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

// Package httpclient impl http client
package httpclient

import (
	"net"
	"sync"
	"time"
)

var (
	defaultDNSCache = NewDNSCache(60 * time.Second)
)

// DNSCache struct
type DNSCache struct {
	// 防止 refresh & lookup 中对 DnsCache.m 的并发访问
	sync.Mutex
	// map[hostname][]IP
	m *sync.Map
	// 刷新 cache 事件间隔
	refreshInterval time.Duration
}

// NewDNSCache 创建 DnsCache
func NewDNSCache(refreshInterval time.Duration) *DNSCache {
	c := &DNSCache{
		m:               new(sync.Map),
		refreshInterval: refreshInterval,
	}
	c.startRefresh()
	return c

}

func (d *DNSCache) startRefresh() {
	go func() {
		for {
			time.Sleep(d.refreshInterval)
			d.Lock()
			d.m = new(sync.Map)
			d.Unlock()
		}
	}()
}

func (d *DNSCache) lookup(host string) ([]net.IP, error) {
	d.Lock()
	m := d.m
	d.Unlock()
	v, ok := m.Load(host)
	if ok {
		return v.([]net.IP), nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	m.Store(host, ips)

	return ips, nil
}
