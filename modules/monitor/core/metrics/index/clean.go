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

package indexmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/olivere/elastic"
)

func (m *IndexManager) startClean() error {
	if int64(m.cfg.IndexCleanInterval) <= 0 {
		return fmt.Errorf("invalid IndexCleanInterval: %v", m.cfg.IndexCleanInterval)
	}
	go func() {
		m.waitAndGetIndices()                                                             // 让 indices 先加载
		time.Sleep(1*time.Second + time.Duration((random.Int63()%10)*int64(time.Second))) // 尽量避免多实例同时进行
		m.log.Infof("enable indices clean, interval: %v", m.cfg.IndexCleanInterval)
		tick := time.Tick(m.cfg.IndexCleanInterval)
		for {
			m.CleanIndices(func(*IndexEntry) bool { return true })
			select {
			case <-tick:
			case req, ok := <-m.clearCh:
				if !ok {
					return
				}
				m.deleteIndices(req.list)
				if req.waitCh != nil {
					close(req.waitCh)
				}
			case <-m.closeCh:
				return
			}
		}
	}()
	return nil
}

// CleanIndices .
func (m *IndexManager) CleanIndices(filter IndexMatcher) error {
	v := m.indices.Load()
	if v == nil {
		return nil
	}
	mc := m.getMetricConfig()
	now := time.Now()
	var removeList []string
	indices := v.(map[string]*indexGroup)
	for _, mg := range indices {
		for _, ng := range mg.Groups {
			for _, entry := range ng.List {
				if filter(entry) && m.needToDelete(entry, mc, now) {
					// atomic.StoreInt32(&entry.Deleted, 1)
					removeList = append(removeList, entry.Index)
				}
			}
			for _, kg := range ng.Groups {
				for _, entry := range kg.List {
					if filter(entry) && m.needToDelete(entry, mc, now) {
						// atomic.StoreInt32(&entry.Deleted, 1)
						removeList = append(removeList, entry.Index)
					}
				}
			}
		}
	}
	if len(removeList) > 0 {
		err := m.deleteIndices(removeList)
		if err != nil {
			return err
		}
		m.toReloadIndices(false)
	}
	return nil
}

func (m *IndexManager) needToDelete(entry *IndexEntry, mc *metricConfig, now time.Time) bool {
	if entry.MaxT.IsZero() || (entry.Num > 0 && entry.Active) {
		return false
	}
	if len(entry.Key) > 0 {
		if mc != nil {
			if d, ok := mc.keysTTL[entry.Key]; ok {
				if int64(d) <= 0 {
					return false
				}
				return now.After(entry.MaxT.Add(d))
			}
		} else {
			return false
		}
	}
	if int64(m.cfg.IndexTTL) <= 0 {
		return false
	}
	return now.After(entry.MaxT.Add(m.cfg.IndexTTL))
}

func (m *IndexManager) deleteIndices(removeList []string) error {
	const size = 10 // 一次性删太多，请求太大会被拒绝
	for len(removeList) >= size {
		err := m.deleteIndex(removeList[:size])
		if err != nil {
			return err
		}
		removeList = removeList[size:]
	}
	if len(removeList) > 0 {
		err := m.deleteIndex(removeList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *IndexManager) deleteIndex(indices []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.RequestTimeout)
	defer cancel()
	resp, err := m.client.DeleteIndex(indices...).Do(ctx)
	if err != nil {
		if e, ok := err.(*elastic.Error); ok {
			if e.Status == 404 {
				return nil
			}
		}
		return err
	}
	if !resp.Acknowledged {
		return fmt.Errorf("delete indices Acknowledged=false")
	}
	m.log.Infof("clean indices %d, %v", len(indices), indices)
	return nil
}
