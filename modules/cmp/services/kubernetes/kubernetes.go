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
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/pkg/k8sclient"
)

var (
	InClusterClient *kubernetes.Clientset
	defaultTimeout  = 2 * time.Second
)

type Kubernetes struct {
	sync.RWMutex
	clients map[string]*k8sclient.K8sClient
}

func (k *Kubernetes) GetCacheClients() map[string]*k8sclient.K8sClient {
	return k.clients
}

func (k *Kubernetes) GetInClusterClient() (*kubernetes.Clientset, error) {
	if InClusterClient != nil {
		return InClusterClient, nil
	}
	rc, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	rc.QPS = 100
	rc.Burst = 100

	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (k *Kubernetes) CreateClient(clusterName string) error {
	nClient, err := k8sclient.NewWithTimeOut(clusterName, defaultTimeout)
	if err != nil {
		return err
	}

	k.writeMap(clusterName, nClient)

	return nil
}

func (k *Kubernetes) GetClient(clusterName string) (*k8sclient.K8sClient, error) {
	client, ok := k.readMap(clusterName)
	if !ok {
		nClient, err := k8sclient.NewWithTimeOut(clusterName, defaultTimeout)
		if err != nil {
			return nil, err
		}
		k.writeMap(clusterName, nClient)
		return nClient, nil
	}
	return client, nil
}

func (k *Kubernetes) UpdateClient(clusterName string) error {
	k.RemoveClient(clusterName)
	return k.CreateClient(clusterName)
}

func (k *Kubernetes) RemoveClient(clusterName string) {
	k.Lock()
	defer k.Unlock()
	delete(k.clients, clusterName)
}

func (k *Kubernetes) readMap(clusterName string) (*k8sclient.K8sClient, bool) {
	k.RLock()
	v, ok := k.clients[clusterName]
	k.RUnlock()
	return v, ok
}

func (k *Kubernetes) writeMap(clusterName string, client *k8sclient.K8sClient) {
	if k.clients == nil {
		k.clients = make(map[string]*k8sclient.K8sClient, 0)
	}

	k.Lock()
	k.clients[clusterName] = client
	k.Unlock()
}
