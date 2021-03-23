package v1alpha1

import (
	rbacv1alpha1_api "istio.io/client-go/pkg/apis/rbac/v1alpha1"
	rbacv1alpha1 "istio.io/client-go/pkg/clientset/versioned/typed/rbac/v1alpha1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func NewRBACClient(addr string) (*rbacv1alpha1.RbacV1alpha1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &rbacv1alpha1_api.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return rbacv1alpha1.New(client), nil
}
