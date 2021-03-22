package v1alpha3

import (
	netv1alpha3_api "istio.io/client-go/pkg/apis/networking/v1alpha3"
	netv1alpha3 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewNetworkingClient creates a new NetworkingV1alpha3 for the given addr.
func NewNetworkingClient(addr string) (*netv1alpha3.NetworkingV1alpha3Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &netv1alpha3_api.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return netv1alpha3.New(client), nil
}
