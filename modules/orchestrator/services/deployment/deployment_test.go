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

package deployment

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type Service struct {
	Value string
}

func TestServer_SomePanic(t *testing.T) {
	source := map[string]*Service{
		"blog": {
			Value: "test",
		},
	}
	target := map[string]*Service{}
	target = nil
	for k, v := range source {
		t, exists := target[k]
		if !exists || t == nil {
			t = &Service{}
			if target == nil {
				target = map[string]*Service{}
			}
			target[k] = t
		}
		target[k].Value = v.Value
	}
	fmt.Println(target)
}

func Test_convertDeploymentRuntimeDTO(t *testing.T) {
	input := apistructs.ServiceGroup{
		Dice: apistructs.Dice{
			Services: []apistructs.Service{
				{
					Name:  "s1",
					Vip:   "v1",
					Ports: []diceyml.ServicePort{{Port: 8080}, {Port: 8081}},
				},
				{
					Name:  "s2",
					Ports: []diceyml.ServicePort{{Port: 9090}, {Port: 9091}},
				},
				{
					Name: "s3",
					Vip:  "v3",
				},
				{
					Name:   "s4",
					Vip:    "v4",
					Ports:  []diceyml.ServicePort{{Port: 80}},
					Labels: map[string]string{"HAPROXY_0_VHOST": "google.com,youtube.com"},
				},
				{
					Name:   "s5",
					Labels: map[string]string{"HAPROXY_0_VHOST": "twitter.com"},
				},
			},
		},
	}
	expected := apistructs.DeploymentStatusRuntimeDTO{
		Services: map[string]*apistructs.DeploymentStatusRuntimeServiceDTO{
			"s1": {
				Host:  "v1",
				Ports: []int{8080, 8081},
			},
			"s2": {
				Ports: []int{9090, 9091},
			},
			"s3": {
				Host: "v3",
			},
		},
		Endpoints: map[string]*apistructs.DeploymentStatusRuntimeServiceDTO{
			"s4": {
				Host:        "v4",
				Ports:       []int{80},
				PublicHosts: []string{"google.com", "youtube.com"},
			},
			"s5": {
				PublicHosts: []string{"twitter.com"},
			},
		},
	}
	output := convertDeploymentRuntimeDTO(&input)

	assert.Equal(t, &expected, output)
}
