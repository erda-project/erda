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

package k8sclient

import (
	"time"

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
	rc, err := GetRestConfig(clusterName)
	if err != nil {
		return nil, err
	}

	return NewForRestConfig(rc, scheme.LocalSchemeBuilder...)
}

// NewWithTimeOut new k8sClient with timeout
func NewWithTimeOut(clusterName string, timeout time.Duration) (*K8sClient, error) {
	rc, err := GetRestConfig(clusterName)
	if err != nil {
		return nil, err
	}

	rc.Timeout = timeout

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

// GetRestConfig get rest config with clusterName
func GetRestConfig(clusterName string) (*rest.Config, error) {
	b := bundle.New(bundle.WithClusterManager())

	ci, err := b.GetCluster(clusterName)
	if err != nil {
		return nil, err
	}

	rc, err := config.ParseManageConfig(clusterName, ci.ManageConfig)
	if err != nil {
		return nil, err
	}

	return rc, nil
}
