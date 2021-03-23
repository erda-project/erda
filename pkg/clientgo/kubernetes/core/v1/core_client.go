package v1

import (
	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewCoreClient creates a new CoreV1Client for the given addr.
func NewCoreClient(addr string) (*corev1client.CoreV1Client, error) {
	config := restclient.GetDefaultConfig("/api")
	config.GroupVersion = &corev1.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return corev1client.New(client), nil
}
