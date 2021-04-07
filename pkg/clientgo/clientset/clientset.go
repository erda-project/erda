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

package clientset

import (
	"k8s.io/client-go/discovery"

	flinkoperatorv1beta1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/flinkoperator/v1beta1"
	openyurtv1alpha1 "github.com/erda-project/erda/pkg/clientgo/clientset/versioned/typed/openyurt/v1alpha1"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	FlinkoperatorV1beta1() flinkoperatorv1beta1.FlinkoperatorV1beta1Interface
	OpenYurtV1Alpha1() openyurtv1alpha1.AppsV1alpha1Interface
}
