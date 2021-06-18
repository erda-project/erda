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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

var (
	once     sync.Once
	clusters []apistructs.ClusterInfo
)

func Initialize(bdl *bundle.Bundle) error {
	var err error
	once.Do(func() {
		clusters, err = bdl.ListClusters("", 0)
		if err != nil {
			return
		}
	})
	return err
}

// GetClusterInfosFirst return all clusters after initialize
// clusters will not change, Just for initial use
func GetClustersFirst() []apistructs.ClusterInfo {
	return clusters
}
