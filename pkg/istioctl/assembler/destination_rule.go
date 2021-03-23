package assembler

import (
	"fmt"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"

	"github.com/erda-project/erda/apistructs"
)

func NewDestinationRule(svc *apistructs.Service) *v1alpha3.DestinationRule {
	result := &v1alpha3.DestinationRule{}
	result.Name = svc.Name
	result.Spec.Host = fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace)
	return result
}
