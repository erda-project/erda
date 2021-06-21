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

package k8sclient

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/k8sclient/config"
	"github.com/erda-project/erda/pkg/k8sclient/scheme"
)

type K8sClient struct {
	ClientSet *kubernetes.Clientset
	CRClient  client.Client
}

// New new K8sClient with clusterName.
func New(clusterName string) (*K8sClient, error) {
	b := bundle.New(bundle.WithClusterManager())

	ci, err := b.GetCluster(clusterName)
	if err != nil {
		return nil, err
	}

	rc, err := config.ParseManageConfig(clusterName, ci.ManageConfig)
	if err != nil {
		return nil, err
	}

	return NewForRestConfig(rc, scheme.LocalSchemeBuilder...)
}

// NewForRestConfig new K8sClient with rest.Config, you can register your custom runtime.Scheme.
func NewForRestConfig(c *rest.Config, schemes ...func(scheme *runtime.Scheme) error) (*K8sClient, error) {
	var kc K8sClient
	var err error

	if kc.ClientSet, err = kubernetes.NewForConfig(c); err != nil {
		return nil, err
	}

	sc := runtime.NewScheme()
	schemeBuilder := &runtime.SchemeBuilder{}

	for _, s := range schemes {
		schemeBuilder.Register(s)
	}

	if err = schemeBuilder.AddToScheme(sc); err != nil {
		return nil, err
	}

	if kc.CRClient, err = client.New(c, client.Options{Scheme: sc}); err != nil {
		return nil, err
	}

	return &kc, nil
}
