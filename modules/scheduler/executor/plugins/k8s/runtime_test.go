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

package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func TestKubernetes_InspectStateful(t *testing.T) {

	kubernetes := &Kubernetes{}

	serviceName := "fake-service"

	sg := &apistructs.ServiceGroup{
		Dice: apistructs.Dice{
			Type: "service",
			ID:   "fakeTest",
			Services: []apistructs.Service{
				{
					Name: "fake-service",
					Ports: []diceyml.ServicePort{
						{
							Port:       1234,
							Protocol:   "HTTPS",
							L4Protocol: "TCP",
						},
						{
							Port:       5678,
							Protocol:   "UDP",
							L4Protocol: "UDP",
						},
					},
				},
			},
		},
	}
	hostName := strutil.Join([]string{serviceName, "service--fakeTest", "svc.cluster.local"}, ".")
	sg, err := kubernetes.inspectStateless(sg)
	assert.Nil(t, err)
	assert.Equal(t, sg.Services[0].ProxyIp, hostName)
	assert.Equal(t, sg.Services[0].Vip, hostName)
	assert.Equal(t, sg.Services[0].ShortVIP, hostName)
	assert.Equal(t, sg.Services[0].ProxyPorts, []int{1234, 5678})
}
