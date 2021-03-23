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
