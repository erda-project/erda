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
	"io/ioutil"
	"time"

	cfgpkg "github.com/recallsong/go-utils/config"
)

func (m *IndexManager) startRollover() error {
	body, err := ioutil.ReadFile(m.cfg.RolloverBodyFile)
	if err != nil {
		return fmt.Errorf("fail to load rollover body file: %s", err)
	}
	body = cfgpkg.EscapeEnv(body)
	m.rolloverBody = string(body)
	if len(m.rolloverBody) <= 0 {
		return fmt.Errorf("invalid RolloverBody")
	}
	if int64(m.cfg.RolloverInterval) <= 0 {
		return fmt.Errorf("invalid RolloverInterval: %v", m.cfg.RolloverInterval)
	}
	m.log.Info("load rollover body: \n", m.rolloverBody)
	go func() {
		m.waitAndGetIndices()                                                             // 让 indices 先加载
		time.Sleep(1*time.Second + time.Duration((random.Int63()%10)*int64(time.Second))) // 尽量让 indices 先加载，并且随机值尽量避免多实例同时进行
		m.log.Infof("enable index rollover, interval: %v", m.cfg.RolloverInterval)
		tick := time.Tick(m.cfg.RolloverInterval)
		for {
			m.RolloverIndices(func(*IndexEntry) bool { return true })
			select {
			case <-tick:
			case <-m.closeCh:
				return
			}
		}
	}()
	return nil
}

// RolloverIndices .
func (m *IndexManager) RolloverIndices(filter IndexMatcher) error {
	return m.rolloverIndices(filter, m.rolloverBody)
}

func (m *IndexManager) rolloverIndices(filter IndexMatcher, body string) error {
	v := m.indices.Load()
	if v == nil {
		return nil
	}
	var num int
	indices := v.(map[string]*indexGroup)
	for metric, mg := range indices {
		for ns, ng := range mg.Groups {
			if len(ng.List) > 0 && ng.List[0].Num > 0 && filter(ng.List[0]) {
				alias := m.indexAlias(metric, ns)
				ok, _ := m.rolloverAlias(alias, body)
				if ok {
					num++
				}
			}
			for key, kg := range ng.Groups {
				if len(kg.List) > 0 && kg.List[0].Num > 0 && filter(kg.List[0]) {
					alias := m.indexAlias(metric, ns+"."+key)
					ok, _ := m.rolloverAlias(alias, body)
					if ok {
						num++
					}
				}
			}
		}
	}
	if num > 0 {
		m.toReloadIndices(false)
	}
	return nil
}

func (m *IndexManager) rolloverAlias(alias, body string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.RequestTimeout)
	defer cancel()
	resp, err := m.client.RolloverIndex(alias).BodyString(body).Do(ctx)
	if err != nil {
		m.log.Errorf("fail to rollover alias %s : %s", alias, err)
		return false, err
	}
	if resp.Acknowledged {
		m.log.Infof("rollover alias %s from %s to %s, %v", alias, resp.OldIndex, resp.NewIndex, resp.Acknowledged)
	} else {
		// m.log.Debugf("rollover alias %s from %s to %s, %v", alias, resp.OldIndex, resp.NewIndex, resp.Acknowledged)
	}
	return resp.Acknowledged, nil
}
