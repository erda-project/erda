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

package clusters

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

// Hook deal k8sClient object in mem
func (c *Clusters) Hook(clusterEvent *apistructs.ClusterEvent) error {
	if clusterEvent == nil {
		return fmt.Errorf("nil clusterEvent object")
	}

	clusterName := clusterEvent.Content.Name

	switch clusterEvent.Action {
	case apistructs.ClusterActionCreate:
		logrus.Debugf("cluster %s action before create, current clients map: %v",
			clusterName, c.k8s.GetCacheClients())
		if err := c.k8s.CreateClient(clusterEvent.Content.Name); err != nil {
			logrus.Errorf("cluster %s action create error: %v", clusterName, err)
			return err
		}
		logrus.Debugf("cluster %s action after create, current clients map: %v",
			clusterName, c.k8s.GetCacheClients())
		return nil
	case apistructs.ClusterActionUpdate:
		logrus.Debugf("cluster %s action before update, current clients map: %v",
			clusterName, c.k8s.GetCacheClients())
		if err := c.k8s.UpdateClient(clusterEvent.Content.Name); err != nil {
			logrus.Errorf("cluster %s action update error: %v", clusterName, err)
			return err
		}
		logrus.Debugf("cluster %s action after update, current clients map: %v",
			clusterName, c.k8s.GetCacheClients())
		return nil
	case apistructs.ClusterActionDelete:
		logrus.Debugf("cluster %s action before delete, current clients map: %v",
			clusterName, c.k8s.GetCacheClients())
		c.k8s.RemoveClient(clusterEvent.Content.Name)
		logrus.Debugf("cluster %s action after delete, current clients map: %v",
			clusterName, c.k8s.GetCacheClients())
		return nil
	default:
		return fmt.Errorf("invaild cluster event action: %s", clusterEvent.Action)
	}
}
