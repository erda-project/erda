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

package scheme

import (
	sparkoperatorv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	elasticsearchv1 "github.com/elastic/cloud-on-k8s/pkg/apis/elasticsearch/v1"
	//flinkoperatoryv1beta1 "github.com/googlecloudplatform/flink-operator/api/v1beta1"
	istioconfigv1alpha2 "istio.io/client-go/pkg/apis/config/v1alpha2"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiorbacv1alpha1 "istio.io/client-go/pkg/apis/rbac/v1alpha1"
	istiosecv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/client-go/kubernetes/scheme"

	openyurtv1alpha1 "github.com/erda-project/erda/pkg/clientgo/apis/openyurt/v1alpha1"
)

// LocalSchemeBuilder register crd scheme
var LocalSchemeBuilder = runtime.SchemeBuilder{
	openyurtv1alpha1.AddToScheme,
	k8sschema.AddToScheme,
	elasticsearchv1.AddToScheme,
	sparkoperatorv1beta2.AddToScheme,
	istioconfigv1alpha2.AddToScheme,
	istionetworkingv1beta1.AddToScheme,
	istionetworkingv1alpha3.AddToScheme,
	istiorbacv1alpha1.AddToScheme,
	istiosecv1beta1.AddToScheme,
	//flinkoperatoryv1beta1.AddToScheme,
	apiextensions.AddToScheme,
}
