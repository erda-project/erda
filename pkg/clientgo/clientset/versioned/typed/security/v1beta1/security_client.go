package v1beta1

import (
	secv1beta1_api "istio.io/client-go/pkg/apis/security/v1beta1"
	secv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1beta1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func NewSecurityClient(addr string) (*secv1beta1.SecurityV1beta1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &secv1beta1_api.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return secv1beta1.New(client), nil
}
