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
	"strings"

	"istio.io/api/networking/v1alpha3"
	"istio.io/api/security/v1beta1"

	"github.com/erda-project/erda/apistructs"
)

func NewPortTlsSettings(svc *apistructs.Service) (drSettings []*v1alpha3.TrafficPolicy_PortTrafficPolicy, paSettings map[uint32]*v1beta1.PeerAuthentication_MutualTLS) {
	paSettings = map[uint32]*v1beta1.PeerAuthentication_MutualTLS{}
	for _, port := range svc.Ports {
		if !strings.EqualFold(port.Protocol, "http") && !strings.EqualFold(port.Protocol, "grpc") {
			drSettings = append(drSettings, &v1alpha3.TrafficPolicy_PortTrafficPolicy{
				Port: &v1alpha3.PortSelector{
					Number: uint32(port.Port),
				},
				Tls: &v1alpha3.ClientTLSSettings{
					Mode: v1alpha3.ClientTLSSettings_DISABLE,
				},
			})
			paSettings[uint32(port.Port)] = &v1beta1.PeerAuthentication_MutualTLS{
				Mode: v1beta1.PeerAuthentication_MutualTLS_DISABLE,
			}
		}
	}
	return
}
