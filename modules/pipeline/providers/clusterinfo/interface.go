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

package clusterinfo

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	ClusterHookApiPath = "/api/pipeline-clusters/actions/hook"
)

// Interface defines the interface for the clusterinfo provider
type Interface interface {
	// GetClusterByName return ClusterInfo query by name, contain cluster connection info and config map data
	GetClusterInfoByName(string) (apistructs.ClusterInfo, error)
	// ListAllClusterInfos return all clusters, contain cluster connection info and config map data
	ListAllClusterInfos() ([]apistructs.ClusterInfo, error)
	// DispatchClusterEvent dispatch cluster event received from eventbox
	// other modules can get event by channel if registered
	DispatchClusterEvent(apistructs.ClusterEvent)
	// BatchUpdateAndDispatchRefresh batch update all clusters, and dispatch refresh event if registered
	BatchUpdateAndDispatchRefresh() error
	// RegisterClusterEvent return channel to receive cluster event
	RegisterClusterEvent() <-chan apistructs.ClusterEvent
	// RegisterRefreshEvent return channel to receive refresh event
	RegisterRefreshEvent() <-chan struct{}
}

func (p *provider) registerClusterHook() error {
	ev := apistructs.CreateHookRequest{
		Name:   "pipeline_watch_cluster_changed",
		Events: []string{bundle.ClusterEvent},
		URL:    strutil.Concat("http://", discover.Pipeline(), ClusterHookApiPath),
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}

	if err := p.bdl.CreateWebhook(ev); err != nil {
		logrus.Errorf("failed to register watch cluster changed event, err: %v", err)
		return err
	}
	return nil
}

// DispatchClusterEvent update the cluster cache then dispatch the event
func (p *provider) DispatchClusterEvent(clusterEvent apistructs.ClusterEvent) {
	switch clusterEvent.Action {
	case apistructs.ClusterActionCreate:
		p.cache.UpdateClusterInfo(clusterEvent.Content)
	case apistructs.ClusterActionUpdate:
		p.cache.UpdateClusterInfo(clusterEvent.Content)
	case apistructs.ClusterActionDelete:
		p.cache.DeleteClusterInfo(clusterEvent.Content.Name)
	default:
		return
	}
	p.notifier.NotifyClusterEvent(clusterEvent)
}

// TriggerManualRefresh update the total cluster cache then dispatch the event
func (p *provider) BatchUpdateAndDispatchRefresh() error {
	if err := p.batchUpdateClusterInfo(); err != nil {
		return err
	}
	p.notifier.NotifyRefreshEvent()
	return nil
}

func (p *provider) batchUpdateClusterInfo() error {
	clusterInfos, err := p.bdl.ListClusters("", 0)
	if err != nil {
		return err
	}
	for _, cluster := range clusterInfos {
		p.cache.UpdateClusterInfo(cluster)
	}
	return nil
}

func (p *provider) continueUpdateAndRefresh() {
	ticker := time.NewTicker(p.Cfg.RefreshClustersInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := p.BatchUpdateAndDispatchRefresh(); err != nil {
				logrus.Errorf("[provider clusterinfo] failed to update cluster info, err: %v", err)
			}
		}
	}
}

func (p *provider) GetClusterInfoByName(clusterName string) (apistructs.ClusterInfo, error) {
	cluster, ok := p.cache.GetClusterInfoByName(clusterName)
	if ok {
		return cluster, nil
	}
	clusterInfo, err := p.bdl.GetCluster(clusterName)
	if err != nil {
		return apistructs.ClusterInfo{}, err
	}
	p.cache.UpdateClusterInfo(*clusterInfo)
	return *clusterInfo, nil
}

// ListAllClusters firstly get all cluster from cache
// if cache is empty, try to get all from bundle and update the cache
func (p *provider) ListAllClusterInfos() ([]apistructs.ClusterInfo, error) {
	clusters := p.cache.GetAllClusters()
	if len(clusters) != 0 {
		return clusters, nil
	}
	if err := p.batchUpdateClusterInfo(); err != nil {
		return nil, err
	}
	return p.cache.GetAllClusters(), nil
}

func (p *provider) RegisterClusterEvent() <-chan apistructs.ClusterEvent {
	return p.notifier.RegisterClusterEvent()
}

func (p *provider) RegisterRefreshEvent() <-chan struct{} {
	return p.notifier.RegisterRefreshEvent()
}

// TODO: GetClusterInfoByName
// after action executor become provider, remove the following methods
func GetClusterInfoByName(clusterName string) (apistructs.ClusterInfo, error) {
	return pd.GetClusterInfoByName(clusterName)
}
