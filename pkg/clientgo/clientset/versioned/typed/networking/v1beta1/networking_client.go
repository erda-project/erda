package v1beta1

import (
	netv1beta1_api "istio.io/client-go/pkg/apis/networking/v1beta1"
	netv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1beta1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewNetworkingClient creates a new NetworkingV1beta1 for the given addr.
func NewNetworkingClient(addr string) (*netv1beta1.NetworkingV1beta1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &netv1beta1_api.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return netv1beta1.New(client), nil
}
