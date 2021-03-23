package v1alpha2

import (
	configv1alpha2_api "istio.io/client-go/pkg/apis/config/v1alpha2"
	configv1alpha2 "istio.io/client-go/pkg/clientset/versioned/typed/config/v1alpha2"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func NewConfigClient(addr string) (*configv1alpha2.ConfigV1alpha2Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &configv1alpha2_api.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return configv1alpha2.New(client), nil
}
