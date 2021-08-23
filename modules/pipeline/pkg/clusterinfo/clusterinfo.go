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
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	ClusterEventListenerLimit = 10
	ClusterHookApiPath        = "/api/pipeline-clusters/actions/hook"
)

var (
	bdl               *bundle.Bundle
	once              sync.Once
	clusterEventChans []chan apistructs.ClusterEvent
	manualTriggerChan = make(chan struct{})
)

func Initialize(bundle *bundle.Bundle) {
	once.Do(func() {
		bdl = bundle
	})
}

func ListAllClusters() ([]apistructs.ClusterInfo, error) {
	return bdl.ListClusters("", 0)
}

// DispatchClusterEvent dispatch every cluster event to registered chan
func DispatchClusterEvent(clusterEvent apistructs.ClusterEvent) {
	for _, ch := range clusterEventChans {
		ch <- clusterEvent
	}
}

func RegisterClusterEvent() (<-chan apistructs.ClusterEvent, error) {
	if len(clusterEventChans) >= ClusterEventListenerLimit {
		return nil, errors.Errorf("number of register cluster event limited, limit num: %d", ClusterEventListenerLimit)
	}
	ch := make(chan apistructs.ClusterEvent, 0)
	clusterEventChans = append(clusterEventChans, ch)
	return ch, nil
}

// RegisterRefreshChan return channel for manual trigger refresh executor
// only for scheduler task manager
func RegisterRefreshChan() <-chan struct{} {
	return manualTriggerChan
}

func TriggerManualRefresh() {
	manualTriggerChan <- struct{}{}
}

// RegisterClusterHook register cluster hook in eventbox
func RegisterClusterHook() error {
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

	if err := bdl.CreateWebhook(ev); err != nil {
		logrus.Errorf("failed to register watch cluster changed event, err: %v", err)
		return err
	}
	return nil
}
