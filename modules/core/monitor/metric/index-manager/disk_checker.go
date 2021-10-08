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
	"sort"
	"time"

	"github.com/olivere/elastic"

	indexloader "github.com/erda-project/erda/modules/core/monitor/metric/index-loader"
)

func (p *provider) runDiskCheckAndClean(ctx context.Context) {
	p.Log.Infof("enable disk clean with interval(%v)", p.Cfg.DiskClean.CheckInterval)
	p.Loader.WaitAndGetIndices(ctx)
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		err := p.checkDiskUsage(ctx)
		if err != nil {
			p.Log.Errorf("failed to check disk: %s", err)
		}
		timer.Reset(p.Cfg.DiskClean.CheckInterval)
	}
}

func (p *provider) getNodeStats() (map[string]*elastic.NodesStatsNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
	defer cancel()
	resp, err := p.ES.Client().NodesStats().Metric("indices", "fs").Do(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Nodes, nil
}

func (p *provider) getClusterState() (*elastic.ClusterStateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
	defer cancel()
	resp, err := p.ES.Client().ClusterState().Metric("routing_nodes").Do(ctx)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// NodeDiskUsage .
type NodeDiskUsage struct {
	ID           string
	Total        int64
	Used         int64
	Store        int64
	UsedPercent  float64
	StorePercent float64
	ExpectDelete int64
	Deleted      int64
}

func (p *provider) getNodeDiskUsage(filter func(*NodeDiskUsage) bool) (map[string]*NodeDiskUsage, error) {
	nodes, err := p.getNodeStats()
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

func (p *provider) getNodeIndices(filter func(*NodeDiskUsage) bool) (map[string]*NodeDiskUsage, map[string]map[string]*NodeDiskUsage, error) {
	nodes, err := p.getNodeDiskUsage(filter)
	if err != nil {
		return nil, nil, err
	}
	if len(nodes) <= 0 {
		return nil, nil, nil
	}
	state, err := p.getClusterState()
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

func (p *provider) checkDiskUsage(ctx context.Context) error {
	nodes, routing, err := p.getNodeIndices(func(n *NodeDiskUsage) bool {
		return n.UsedPercent >= p.Cfg.DiskClean.HighDiskUsagePercent &&
			n.StorePercent >= p.Cfg.DiskClean.MinIndicesStorePercent &&
			n.Store >= p.minIndicesStoreInDisk
	})
	if err != nil {
		return err
	}
	for {
		if len(nodes) <= 0 || len(routing) <= 0 {
			return nil
		}

		// estimate the number of deletes
		for _, n := range nodes {
			targetDiskUsage := int64(float64(p.Cfg.DiskClean.LowDiskUsagePercent) / 100 * float64(n.Total))
			delBytes := n.Used - targetDiskUsage
			minStore := int64(float64(n.Total) * p.Cfg.DiskClean.MinIndicesStorePercent / 100)
			if p.minIndicesStoreInDisk > minStore {
				minStore = p.minIndicesStoreInDisk
			}
			// clean up the index under the premise of ensuring minimum index storage
			if delBytes > minStore {
				delBytes = delBytes - minStore
			}
			n.ExpectDelete = delBytes
		}

		// find indices to clean
		_, sortedIndices := p.getSortedIndices()
		var removeList []string
		for _, entry := range sortedIndices {
			for _, indices := range routing {
				if node, ok := indices[entry.Index]; ok {
					if node.Deleted > node.ExpectDelete {
						continue
					}
					node.Deleted += entry.StoreBytes
					removeList = append(removeList, entry.Index)
				}
			}
		}
		if len(removeList) <= 0 && len(routing) > 0 && len(nodes) > 0 {
			break
		}

		// delete indices
		req := &clearRequest{
			list:   removeList,
			waitCh: make(chan struct{}),
		}
		select {
		case <-ctx.Done():
			return nil
		case p.clearCh <- req:
		}
		// wait for index deletion to complete
		select {
		case <-ctx.Done():
			return nil
		case <-req.waitCh:
		}

		p.Loader.ReloadIndices()

		// restore the node whose disk capacity was previously overloaded
		nodes, routing, err = p.getNodeIndices(func(n *NodeDiskUsage) bool {
			return nodes[n.ID] != nil &&
				n.UsedPercent >= p.Cfg.DiskClean.LowDiskUsagePercent &&
				n.StorePercent >= p.Cfg.DiskClean.MinIndicesStorePercent &&
				n.Store >= p.minIndicesStoreInDisk
		})
		if err != nil {
			return err
		}
	}
	if p.Cfg.EnableRollover {
		_, sortedIndices := p.getSortedRolloverIndices()
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
					alias := p.indexAlias(entry.Metric, ns)
					ok, _ := p.rolloverAlias(alias, p.rolloverBodyForDiskClean)
					if ok {
						node.Deleted += entry.StoreBytes
						removeList = append(removeList, entry.Index)
					}
				}
			}
		}
		if len(removeList) >= 0 {
			// delete indices
			req := &clearRequest{
				list:   removeList,
				waitCh: make(chan struct{}),
			}
			select {
			case <-ctx.Done():
				return nil
			case p.clearCh <- req:
			}
			// wait for index deletion to complete
			select {
			case <-ctx.Done():
				return nil
			case <-req.waitCh:
			}

			p.Loader.ReloadIndices()
		} else {
			p.Log.Warnf("high disk usage, but not find indices to delete")
		}
	}
	return nil
}

func (p *provider) getSortedIndices() (map[string]*indexloader.IndexGroup, []*indexloader.IndexEntry) {
	indices := p.Loader.AllIndices()
	var sortedIndices []*indexloader.IndexEntry
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
	// ascending by maximum time and size
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

func (p *provider) getSortedRolloverIndices() (map[string]*indexloader.IndexGroup, []*indexloader.IndexEntry) {
	indices := p.Loader.AllIndices()
	var sortedIndices []*indexloader.IndexEntry
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
