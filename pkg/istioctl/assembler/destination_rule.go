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
