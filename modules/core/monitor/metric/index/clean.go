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
		m.waitAndGetIndices()                                                          // Let the indices load first
		time.Sleep(1*time.Second + time.Duration(random.Int63n(9)*int64(time.Second))) // Try to avoid multiple instances at the same time
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
	const size = 10 // Delete too much at once and the request will be rejected
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
