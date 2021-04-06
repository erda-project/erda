// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
