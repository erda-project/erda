package assembler

import (
	securityv1beta1 "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	"istio.io/client-go/pkg/apis/security/v1beta1"

	"github.com/erda-project/erda/apistructs"
)

func NewPeerAuthentication(svc *apistructs.Service) *v1beta1.PeerAuthentication {
	result := &v1beta1.PeerAuthentication{}
	result.Name = svc.Name
	result.Spec.Mtls = &securityv1beta1.PeerAuthentication_MutualTLS{
		Mode: securityv1beta1.PeerAuthentication_MutualTLS_STRICT,
	}
	result.Spec.Selector = &typev1beta1.WorkloadSelector{
		MatchLabels: map[string]string{
			"app": svc.Name,
		},
	}
	return result
}
