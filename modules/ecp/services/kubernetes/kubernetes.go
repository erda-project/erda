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

	"github.com/erda-project/erda/pkg/k8sclient"
)

var k8sClients = make(map[string]*k8sclient.K8sClient, 0)

type Kubernetes struct{}

func New() *Kubernetes {
	return &Kubernetes{}
}

// getK8sClient Get or create kubernetes client by cluster name.
func (k *Kubernetes) getK8sClient(clusterName string) (*k8sclient.K8sClient, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("empty cluster name")
	}

	if k8sClient, ok := k8sClients[clusterName]; ok {
		return k8sClient, nil
	}

	k8sClient, err := k8sclient.New(clusterName)
	if err != nil {
		errMsg := fmt.Errorf("cluster %s k8sclient create error, error: %+v", clusterName, err)
		logrus.Error(errMsg)
		return nil, errMsg
	}

	k8sClients[clusterName] = k8sClient

	return k8sClient, nil
}
