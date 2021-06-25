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

package clusterinfo

import (
	"sync"

	"github.com/gogap/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

const (
	ClusterEventListenerLimit = 10
)

var (
	once              sync.Once
	clusters          []apistructs.ClusterInfo
	clusterEventChans []chan apistructs.ClusterEvent
)

func Initialize(bdl *bundle.Bundle) error {
	var err error
	once.Do(func() {
		clusterEventChans = []chan apistructs.ClusterEvent{}
		clusters, err = bdl.ListClusters("", 0)
		if err != nil {
			return
		}
	})
	return err
}

// GetClusterInfosFirst return all clusters after initialize
// clusters will not change, Just for initial use
func GetClustersInitialize() []apistructs.ClusterInfo {
	return clusters
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
