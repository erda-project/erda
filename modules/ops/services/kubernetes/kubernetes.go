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

package kubernetes

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/clientgo"
)

var (
	clusterInfos = make(map[string]*apistructs.ClusterInfo, 0)
	clientSets   = make(map[string]*clientgo.ClientSet, 0)
)

type Kubernetes struct {
	bdl *bundle.Bundle
}

type Option func(*Kubernetes)

func New(options ...Option) *Kubernetes {
	r := &Kubernetes{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithBundle With bundle.
func WithBundle(bdl *bundle.Bundle) Option {
	return func(k *Kubernetes) {
		k.bdl = bdl
	}
}

// getClusterInfo Get cluster info from cache.
func (k *Kubernetes) getClusterInfo(clusterName string) (*apistructs.ClusterInfo, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("empty cluster name")
	}

	if clusterInfo, ok := clusterInfos[clusterName]; ok {
		return clusterInfo, nil
	}
	clusterInfo, err := k.bdl.GetCluster(clusterName)
	if err != nil {
		return nil, fmt.Errorf("query cluster info failed, cluster:%s, err:%v", clusterName, err)
	}
	clusterInfos[clusterName] = clusterInfo
	return clusterInfo, nil
}

// getClientSet Get or create client-go client by cluster name.
func (k *Kubernetes) getClientSet(clusterName string) (*clientgo.ClientSet, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("empty cluster name")
	}

	if clientSet, ok := clientSets[clusterName]; ok {
		return clientSet, nil
	}

	clusterInfo, err := k.getClusterInfo(clusterName)
	if err != nil {
		return nil, err
	}

	if clusterInfo.SchedConfig == nil || clusterInfo.SchedConfig.MasterURL == "" {
		return nil, fmt.Errorf("empty inet address, cluster:%s", clusterName)
	}

	clientSet, err := clientgo.New(clusterInfo.SchedConfig.MasterURL)
	if err != nil {
		logrus.Errorf("cluster %s clientset create error, parse master url: %s, error: %+v", clusterName, clusterInfo.SchedConfig.MasterURL, err)
		return nil, fmt.Errorf("cluster %s clientset create error", clusterName)
	}

	clientSets[clusterName] = clientSet
	return clientSet, nil
}
