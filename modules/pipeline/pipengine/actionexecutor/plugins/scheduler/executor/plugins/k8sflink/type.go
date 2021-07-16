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

package k8sflink

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/pkg/k8sclient"
)

type K8sFlink struct {
	name        types.Name
	client      *k8sclient.K8sClient
	clusterName string
	cluster     apistructs.ClusterInfo
}

func New(name types.Name, clusterName string, cluster apistructs.ClusterInfo) (*K8sFlink, error) {
	k, err := k8sclient.New(clusterName)
	if err != nil {
		return nil, err
	}
	return &K8sFlink{name: name, client: k, clusterName: clusterName, cluster: cluster}, nil
}

func (k *K8sFlink) Kind() types.Kind {
	return Kind
}

func (k *K8sFlink) Name() types.Name {
	return k.name
}
