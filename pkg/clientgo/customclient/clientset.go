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

package customclient

import (
	configv1alpha2 "istio.io/client-go/pkg/clientset/versioned/typed/config/v1alpha2"
	netv1alpha3 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"
	netv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1beta1"
	rbacv1alpha1 "istio.io/client-go/pkg/clientset/versioned/typed/rbac/v1alpha1"
	secv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1beta1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	apiextapiv1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/apiextensions/v1"
	erdaconfigv1alpha2 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/config/v1alpha2"
	elasticsearchv1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/elasticsearch/v1"
	flinkoperatorv1beta1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/flinkoperator/v1beta1"
	erdanetworkingv1alpha3 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/networking/v1alpha3"
	erdanetworkingv1beta1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/networking/v1beta1"
	openYurtV1alpha1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/openyurt/v1alpha1"
	erdarbacv1alpha1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/rbac/v1alpha1"
	redisfailover "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/redisfailover/v1"
	redisfailoverv1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/redisfailover/v1"
	erdasecurityv1beta1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/security/v1beta1"
	sparkoperatorv1beta1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/sparkoperator/v1beta1"
	sparkoperatorv1beta2 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/sparkoperator/v1beta2"
	erdadiscovery "github.com/erda-project/erda/pkg/clientgo/discovery"
)

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	flinkoperatorV1beta1 *flinkoperatorv1beta1.FlinkoperatorV1beta1Client
	sparkoperatorV1beta1 *sparkoperatorv1beta1.SparkoperatorV1beta1Client
	sparkoperatorV1beta2 *sparkoperatorv1beta2.SparkoperatorV1beta2Client
	openYurtV1alpha1     *openYurtV1alpha1.AppsV1alpha1Client
	configV1alpha2       *configv1alpha2.ConfigV1alpha2Client
	networkingV1alpha3   *netv1alpha3.NetworkingV1alpha3Client
	networkingV1beta1    *netv1beta1.NetworkingV1beta1Client
	rbacV1alpha1         *rbacv1alpha1.RbacV1alpha1Client
	securityV1beta1      *secv1beta1.SecurityV1beta1Client
	redisfailoverV1      *redisfailover.DatabasesV1Client
	elasticsearchV1      *elasticsearchv1.ElasticsearchV1Client
	apiExtensionsV1      *apiextv1.ApiextensionsV1Client
}

// FlinkoperatorV1beta1 retrieves the FlinkoperatorV1beta1Client
func (c *Clientset) FlinkoperatorV1beta1() flinkoperatorv1beta1.FlinkoperatorV1beta1Interface {
	return c.flinkoperatorV1beta1
}

// SparkoperatorV1beta1 retrieves the SparkoperatorV1beta1Interface
func (c *Clientset) SparkoperatorV1beta1() sparkoperatorv1beta1.SparkoperatorV1beta1Interface {
	return c.sparkoperatorV1beta1
}

// SparkoperatorV1beta1 retrieves the SparkoperatorV1beta2Interface
func (c *Clientset) SparkoperatorV1beta2() sparkoperatorv1beta2.SparkoperatorV1beta2Interface {
	return c.sparkoperatorV1beta2
}

// Open retrieves the OpenYurt AppsV1alpha1Client
func (c *Clientset) OpenYurtV1alpha1() openYurtV1alpha1.AppsV1alpha1Interface {
	return c.openYurtV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

func (c *Clientset) ConfigV1alpha2() configv1alpha2.ConfigV1alpha2Interface {
	return c.configV1alpha2
}

func (c *Clientset) NetworkingV1alpha3() netv1alpha3.NetworkingV1alpha3Interface {
	return c.networkingV1alpha3
}

func (c *Clientset) NetworkingV1beta1() netv1beta1.NetworkingV1beta1Interface {
	return c.networkingV1beta1
}

func (c *Clientset) RbacV1alpha1() rbacv1alpha1.RbacV1alpha1Interface {
	return c.rbacV1alpha1
}

func (c *Clientset) SecurityV1beta1() secv1beta1.SecurityV1beta1Interface {
	return c.securityV1beta1
}

// SparkoperatorV1beta1 retrieves the SparkoperatorV1beta2Interface
func (c *Clientset) RedisfailoverV1() redisfailover.DatabasesV1Interface {
	return c.redisfailoverV1
}

// ElasticsearchV1 retrieves the ElasticsearchV1Interface
func (c *Clientset) ElasticsearchV1() elasticsearchv1.ElasticsearchV1Interface {
	return c.elasticsearchV1
}

// ApiExtensionsV1 retrieves the ApiextensionsV1Interface
func (c *Clientset) ApiExtensionsV1() apiextv1.ApiextensionsV1Interface {
	return c.apiExtensionsV1
}

// NewCustomClientSet creates a new Clientset for the given addr.
func NewCustomClientSet(addr string) (*Clientset, error) {
	var cs Clientset
	var err error
	cs.flinkoperatorV1beta1, err = flinkoperatorv1beta1.NewFlinkOpeartorClient(addr)
	if err != nil {
		return nil, err
	}

	cs.sparkoperatorV1beta1, err = sparkoperatorv1beta1.NewSparkOpeartorClient(addr)
	if err != nil {
		return nil, err
	}

	cs.sparkoperatorV1beta2, err = sparkoperatorv1beta2.NewSparkOpeartorClient(addr)
	if err != nil {
		return nil, err
	}

	cs.openYurtV1alpha1, err = openYurtV1alpha1.NewOpenYurtClient(addr)
	if err != nil {
		return nil, err
	}

	cs.networkingV1alpha3, err = erdanetworkingv1alpha3.NewNetworkingClient(addr)
	if err != nil {
		return nil, err
	}

	cs.networkingV1beta1, err = erdanetworkingv1beta1.NewNetworkingClient(addr)
	if err != nil {
		return nil, err
	}

	cs.rbacV1alpha1, err = erdarbacv1alpha1.NewRBACClient(addr)
	if err != nil {
		return nil, err
	}

	cs.configV1alpha2, err = erdaconfigv1alpha2.NewConfigClient(addr)
	if err != nil {
		return nil, err
	}

	cs.securityV1beta1, err = erdasecurityv1beta1.NewSecurityClient(addr)
	if err != nil {
		return nil, err
	}

	cs.redisfailoverV1, err = redisfailoverv1.NewRedisFailoverClient(addr)
	if err != nil {
		return nil, err
	}

	cs.elasticsearchV1, err = elasticsearchv1.NewElasticsearchClient(addr)
	if err != nil {
		return nil, err
	}

	cs.apiExtensionsV1, err = apiextapiv1.NewApiExtensionsClient(addr)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = erdadiscovery.NewDiscoveryClient(addr)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewCustomClientSetWithConfig creates a new Clientset for the given restConfig.
func NewCustomClientSetWithConfig(restConfig *rest.Config) (*Clientset, error) {
	var (
		cs  Clientset
		err error
	)

	cs.flinkoperatorV1beta1, err = flinkoperatorv1beta1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.sparkoperatorV1beta1, err = sparkoperatorv1beta1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.sparkoperatorV1beta2, err = sparkoperatorv1beta2.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.openYurtV1alpha1, err = openYurtV1alpha1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.networkingV1alpha3, err = erdanetworkingv1alpha3.NewNetworkingClientWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.networkingV1beta1, err = erdanetworkingv1beta1.NewNetworkingClientWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.rbacV1alpha1, err = erdarbacv1alpha1.NewRBACClientWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.configV1alpha2, err = erdaconfigv1alpha2.NewConfigClientWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.securityV1beta1, err = erdasecurityv1beta1.NewSecurityClientWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.redisfailoverV1, err = redisfailoverv1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.elasticsearchV1, err = elasticsearchv1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.apiExtensionsV1, err = apiextv1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = erdadiscovery.NewDiscoveryClientWithConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
