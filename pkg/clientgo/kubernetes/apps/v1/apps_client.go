package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewAppClient creates a new AppsV1Client for the given addr.
func NewAppClient(addr string) (*appsv1client.AppsV1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &appsv1.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return appsv1client.New(client), nil
}
