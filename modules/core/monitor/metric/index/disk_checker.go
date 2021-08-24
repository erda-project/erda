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
	"io/ioutil"
	"sort"
	"time"

	"github.com/olivere/elastic"
	cfgpkg "github.com/recallsong/go-utils/config"
	"github.com/recallsong/go-utils/lang/size"

	mutex "github.com/erda-project/erda-infra/providers/etcd-mutex"
)

func (m *IndexManager) getNodeStats() (map[string]*elastic.NodesStatsNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.RequestTimeout)
	defer cancel()
	resp, err := m.client.NodesStats().Metric("indices", "fs").Do(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Nodes, nil
}

func (m *IndexManager) getClusterState() (*elastic.ClusterStateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.RequestTimeout)
	defer cancel()
	resp, err := m.client.ClusterState().Metric("routing_nodes").Do(ctx)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// NodeDiskUsage .
type NodeDiskUsage struct {
	ID    string
	Total int64 // Total disk capacity of the node
	Used  int64 // The node is already using disk capacity
	Store int64 // Index storage

	UsedPercent  float64 // The node is already using a percentage
	StorePercent float64 // Index storage percentage

	ExpectDelete int64 // Storage expected to be cleared
	Deleted      int64 // The amount of memory currently cleared
}

func (m *IndexManager) getNodeDiskUsage(filter func(*NodeDiskUsage) bool) (map[string]*NodeDiskUsage, error) {
	nodes, err := m.getNodeStats()
	if err != nil {
		return nil, err
	}
	diskUsage := make(map[string]*NodeDiskUsage)
	for id, node := range nodes {
		usage := &NodeDiskUsage{
			ID:           id,
			Total:        node.FS.Total.TotalInBytes,
			Used:         node.FS.Total.TotalInBytes - node.FS.Total.AvailableInBytes,
			Store:        node.Indices.Store.SizeInBytes,
			UsedPercent:  float64(node.FS.Total.TotalInBytes-node.FS.Total.AvailableInBytes) / float64(node.FS.Total.TotalInBytes) * 100,
			StorePercent: float64(node.Indices.Store.SizeInBytes) / float64(node.FS.Total.TotalInBytes) * 100,
		}
		if filter == nil || filter(usage) {
			diskUsage[id] = usage
		}
	}
	return diskUsage, nil
}

func (m *IndexManager) getNodeIndices(filter func(*NodeDiskUsage) bool) (map[string]*NodeDiskUsage, map[string]map[string]*NodeDiskUsage, error) {
	nodes, err := m.getNodeDiskUsage(filter)
	if err != nil {
		return nil, nil, err
	}
	if len(nodes) <= 0 {
		return nil, nil, nil
	}
	state, err := m.getClusterState()
	if err != nil {
		return nodes, nil, err
	}
	routing := make(map[string]map[string]*NodeDiskUsage)
	for id, node := range nodes {
		indices := routing[id]
		if indices == nil {
			indices = make(map[string]*NodeDiskUsage)
			routing[id] = indices
		}
		shards := state.RoutingNodes.Nodes[id]
		for _, shard := range shards {
			if _, ok := indices[shard.Index]; !ok {
				indices[shard.Index] = node
			}
		}
	}
	return nodes, routing, nil
}

type clearRequest struct {
	waitCh chan struct{}
	list   []string
}

func (m *IndexManager) checkDiskUsage() error {
	nodes, routing, err := m.getNodeIndices(func(n *NodeDiskUsage) bool {
		return n.UsedPercent >= m.cfg.DiskClean.HighDiskUsagePercent &&
			n.StorePercent >= m.cfg.DiskClean.MinIndicesStorePercent &&
			n.Store >= m.cfg.DiskClean.minIndicesStore
	})
	if err != nil {
		return err
	}
	for {
		if len(nodes) <= 0 || len(routing) <= 0 {
			// 达到要求，无需再清理
			return nil
		}
		// Estimate the number of deletes
		for _, n := range nodes {
			// Amount of expected cleanup
			delBytes := int64(float64(n.Used) - float64(m.cfg.DiskClean.LowDiskUsagePercent)/100*float64(n.Total))
			// The minimum index usage guaranteed not to be dropped
			minStore := int64(float64(n.Total) * m.cfg.DiskClean.MinIndicesStorePercent / 100)
			if m.cfg.DiskClean.minIndicesStore > minStore {
				minStore = m.cfg.DiskClean.minIndicesStore
			}
			// Clean up the index under the premise of ensuring minimum index storage
			if delBytes > minStore {
				delBytes = delBytes - minStore
			}
			n.ExpectDelete = delBytes
		}
		_, sortedIndices := m.getSortedIndices()
		var removeList []string
		for _, entry := range sortedIndices {
			for _, indices := range routing {
				if node, ok := indices[entry.Index]; ok {
					if node.Deleted > node.ExpectDelete {
						continue
					}
					node.Deleted += entry.StoreSize
					removeList = append(removeList, entry.Index)
				}
			}
		}
		if len(removeList) <= 0 && len(routing) > 0 && len(nodes) > 0 {
			// There are nodes that need to be cleaned up, but no indexes that need to be dropped, pass the rollover and then delete
			break
		}
		req := &clearRequest{
			list:   removeList,
			waitCh: make(chan struct{}),
		}
		m.clearCh <- req
		<-req.waitCh            // Wait for index deletion to complete
		m.toReloadIndices(true) // Wait for the index to reload
		// Restore the node whose disk capacity was previously overloaded
		nodes, routing, err = m.getNodeIndices(func(n *NodeDiskUsage) bool {
			return nodes[n.ID] != nil &&
				n.UsedPercent >= m.cfg.DiskClean.LowDiskUsagePercent &&
				n.StorePercent >= m.cfg.DiskClean.MinIndicesStorePercent &&
				n.Store >= m.cfg.DiskClean.minIndicesStore
		})
		if err != nil {
			return err
		}
	}
	if m.cfg.EnableRollover {
		_, sortedIndices := m.getSortedRolloverIndices()
		var removeList []string
		for _, entry := range sortedIndices {
			for _, indices := range routing {
				if node, ok := indices[entry.Index]; ok {
					if node.Deleted > node.ExpectDelete {
						continue
					}
					ns := entry.Namespace
					if len(entry.Key) > 0 {
						ns = ns + "." + entry.Key
					}
					alias := m.indexAlias(entry.Metric, ns)
					ok, _ := m.rolloverAlias(alias, m.rolloverBodyForDiskClean)
					if ok {
						node.Deleted += entry.StoreSize
						removeList = append(removeList, entry.Index)
					}
				}
			}
		}
		if len(removeList) >= 0 {
			req := &clearRequest{
				list:   removeList,
				waitCh: make(chan struct{}),
			}
			m.clearCh <- req
			<-req.waitCh
			m.toReloadIndices(true)
		} else {
			m.log.Warnf("high disk usage, but not find indices to delete")
		}
	}
	return nil
}

func (m *IndexManager) startDiskCheck(lock mutex.Mutex) error {
	if m.cfg.EnableRollover {
		body, err := ioutil.ReadFile(m.cfg.DiskClean.RolloverBodyFile)
		if err != nil {
			return fmt.Errorf("fail to load rollover body file for disk: %s", err)
		}
		body = cfgpkg.EscapeEnv(body)
		m.rolloverBodyForDiskClean = string(body)
		if len(m.rolloverBodyForDiskClean) <= 0 {
			return fmt.Errorf("invalid RolloverBody for disk clean")
		}
		m.log.Info("load rollover body for disk clean: \n", m.rolloverBodyForDiskClean)
	}
	if int64(m.cfg.DiskClean.CheckInterval) <= 0 {
		return fmt.Errorf("invalid DiskClean.CheckInterval: %v", m.cfg.DiskClean.CheckInterval)
	}
	minIndicesStore, err := size.ParseBytes(m.cfg.DiskClean.MinIndicesStore)
	if err != nil {
		return fmt.Errorf("invalid min_indices_store: %s", err)
	}
	m.cfg.DiskClean.minIndicesStore = minIndicesStore
	go func() {
		if lock != nil {
			defer lock.Close()
		}
		m.waitAndGetIndices()                                                           // Let the indices load first
		time.Sleep(10*time.Second + time.Duration(random.Int63n(9)*int64(time.Second))) // Try to avoid multiple instances at the same time
		m.log.Infof("enable disk clean, interval: %v", m.cfg.DiskClean.CheckInterval)
		for {
			if lock != nil {
				err := lock.Lock(context.Background())
				if err == nil {
					err = m.checkDiskUsage()
					if err != nil {
						m.log.Errorf("fail to check disk: %s", err)
					}
				}
				lock.Unlock(context.Background())
			} else {
				err := m.checkDiskUsage()
				if err != nil {
					m.log.Errorf("fail to check disk: %s", err)
				}
			}
			select {
			case <-time.After(m.cfg.DiskClean.CheckInterval):
			case <-m.closeCh:
				return
			}
		}
	}()
	return nil
}

func (m *IndexManager) getSortedIndices() (map[string]*indexGroup, []*IndexEntry) {
	v := m.indices.Load()
	if v == nil {
		return nil, nil
	}
	indices := v.(map[string]*indexGroup)
	var sortedIndices []*IndexEntry
	for _, mg := range indices {
		for _, ng := range mg.Groups {
			for i := len(ng.List) - 1; i >= 0; i-- {
				entry := ng.List[i]
				if entry.Num > 0 && i == 0 {
					break
				}
				sortedIndices = append(sortedIndices, entry)
			}
			for _, kg := range ng.Groups {
				for i := len(kg.List) - 1; i >= 0; i-- {
					entry := kg.List[i]
					if entry.Num > 0 && i == 0 {
						break
					}
					sortedIndices = append(sortedIndices, entry)
				}
			}
		}
	}
	// Ascending by maximum time and size
	sort.Slice(sortedIndices, func(i, j int) bool {
		a, b := sortedIndices[i], sortedIndices[j]
		at, bt := a.MaxT.Truncate(time.Hour), b.MaxT.Truncate(time.Hour)
		if at.Equal(bt) {
			return a.StoreSize < b.StoreSize
		}
		return at.Before(bt)
	})
	return indices, sortedIndices
}

func (m *IndexManager) getSortedRolloverIndices() (map[string]*indexGroup, []*IndexEntry) {
	v := m.indices.Load()
	if v == nil {
		return nil, nil
	}
	indices := v.(map[string]*indexGroup)
	var sortedIndices []*IndexEntry
	for _, mg := range indices {
		for _, ng := range mg.Groups {
			if len(ng.List) == 1 && ng.List[0].Num > 0 {
				sortedIndices = append(sortedIndices, ng.List[0])
			}
			for _, kg := range ng.Groups {
				if len(kg.List) == 1 && kg.List[0].Num > 0 {
					sortedIndices = append(sortedIndices, kg.List[0])
				}
			}
		}
	}
	sort.Slice(sortedIndices, func(i, j int) bool {
		a, b := sortedIndices[i], sortedIndices[j]
		return a.StoreSize < b.StoreSize
	})
	return indices, sortedIndices
}
