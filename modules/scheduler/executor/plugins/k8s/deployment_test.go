// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/secret"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestNewDeployment(t *testing.T) {
	service := &apistructs.Service{
		Name:          "test-service",
		Namespace:     "test",
		Image:         "test",
		ImageUsername: "",
		ImagePassword: "",
		Cmd:           "",
		Ports:         nil,
		ProxyPorts:    nil,
		Vip:           "",
		ShortVIP:      "",
		ProxyIp:       "",
		PublicIp:      "",
		Scale:         0,
		Resources: apistructs.Resources{
			Cpu: 0.1,
			Mem: 512,
		},
		Depends:            nil,
		Env:                nil,
		Labels:             nil,
		DeploymentLabels:   nil,
		Selectors:          nil,
		Binds:              nil,
		Volumes:            nil,
		Hosts:              nil,
		HealthCheck:        nil,
		NewHealthCheck:     nil,
		SideCars:           nil,
		InitContainer:      nil,
		InstanceInfos:      nil,
		MeshEnable:         nil,
		TrafficSecurity:    diceyml.TrafficSecurity{},
		WorkLoad:           "",
		ProjectServiceName: "",
		K8SSnippet:         nil,
		StatusDesc:         apistructs.StatusDesc{},
	}

	servicegroup := &apistructs.ServiceGroup{
		Dice: apistructs.Dice{
			ID:                   "test",
			Type:                 "service",
			Labels:               map[string]string{},
			Services:             []apistructs.Service{*service},
			ServiceDiscoveryKind: "",
			ServiceDiscoveryMode: "",
			ProjectNamespace:     "",
		},
	}

	k := &Kubernetes{
		secret: &secret.Secret{},
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(k.secret), "Get", func(sec *secret.Secret, namespace, name string) (*apiv1.Secret, error) {
		return &apiv1.Secret{}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(k), "CopyErdaSecrets", func(kube *Kubernetes, originns, dstns string) ([]apiv1.Secret, error) {
		return []apiv1.Secret{}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(k), "AddPodMountVolume", func(kube *Kubernetes, service *apistructs.Service, podSpec *apiv1.PodSpec,
		secretvolmounts []apiv1.VolumeMount, secretvolumes []apiv1.Volume) error {
		return nil
	})
	_, err := k.newDeployment(service, servicegroup)
	assert.Equal(t, err, nil)
}

func TestUpdateContainerResourceEnv(t *testing.T) {
	k := Kubernetes{}

	container := apiv1.Container{
		Name:       "",
		Image:      "",
		Command:    nil,
		Args:       nil,
		WorkingDir: "",
		Ports:      nil,
		EnvFrom:    nil,
		Env: []apiv1.EnvVar{
			{
				Name:  "DICE_CPU_ORIGIN",
				Value: "0.100000",
			},
			{
				Name:  "DICE_CPU_REQUEST",
				Value: "0.010",
			},
			{
				Name:  "DICE_CPU_LIMIT",
				Value: "0.100",
			},
			{
				Name:  "DICE_MEM_ORIGIN",
				Value: "1024.000000",
			},
			{
				Name:  "DICE_MEM_REQUEST",
				Value: "1024",
			},
			{
				Name:  "DICE_MEM_LIMIT",
				Value: "1024",
			},
		},
		Resources: apiv1.ResourceRequirements{
			Limits: apiv1.ResourceList{
				"cpu":    resource.MustParse("100m"),
				"memory": resource.MustParse("1024Mi"),
			},
			Requests: apiv1.ResourceList{
				"cpu":    resource.MustParse("10m"),
				"memory": resource.MustParse("1024Mi"),
			},
		},
	}
	originResource := apistructs.Resources{Cpu: 0.1, Mem: 1024}
	k.UpdateContainerResourceEnv(originResource, &container)
	assert.Equal(t, container.Env[0].Value, "0.100000")
	assert.Equal(t, container.Env[5].Value, "1024")
}
