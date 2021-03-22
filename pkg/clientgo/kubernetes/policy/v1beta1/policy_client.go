package v1beta1

import (
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	policyv1beta1client "k8s.io/client-go/kubernetes/typed/policy/v1beta1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewPolicyClient creates a new PolicyV1beta1 for the given addr.
func NewPolicyClient(addr string) (*policyv1beta1client.PolicyV1beta1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &policyv1beta1.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return policyv1beta1client.New(client), nil
}
