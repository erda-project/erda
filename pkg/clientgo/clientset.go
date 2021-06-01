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

package clientgo

import (
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/pkg/clientgo/customclient"
	"github.com/erda-project/erda/pkg/clientgo/kubernetes"
)

type ClientSet struct {
	K8sClient    *kubernetes.Clientset
	CustomClient *customclient.Clientset
}

func New(addr string) (*ClientSet, error) {
	var cs ClientSet
	var err error
	cs.K8sClient, err = kubernetes.NewKubernetesClientSet(addr)
	if err != nil {
		return nil, err
	}
	cs.CustomClient, err = customclient.NewCustomClientSet(addr)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

func NewClientSet(restConfig *rest.Config) (*ClientSet, error) {
	var (
		cs  ClientSet
		err error
	)

	cs.K8sClient, err = kubernetes.NewKubernetesClientSetWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.CustomClient, err = customclient.NewCustomClientSetWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
