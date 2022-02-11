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
	appsv1 "k8s.io/api/apps/v1"
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
	deploy, err := k.newDeployment(service, servicegroup)
	assert.Equal(t, err, nil)
	k.setDeploymentZeroReplica(deploy)
	assert.Equal(t, *deploy.Spec.Replicas, int32(0))
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

func TestSetPodAnnotationsBaseContainerEnvs(t *testing.T) {
	type args struct {
		container      apiv1.Container
		podAnnotations map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test_01",
			args: args{
				container: apiv1.Container{
					Env: []apiv1.EnvVar{
						{
							Name:  "DICE_ORG_NAME",
							Value: "xxx",
						},
						{
							Name:  "DICE_CLUSTER_NAME",
							Value: "xxx",
						},
						{
							Name:  "DICE_PROJECT_NAME",
							Value: "xxx",
						},
						{
							Name:  "DICE_SERVICE_NAME",
							Value: "xxx",
						},
						{
							Name:  "DICE_APPLICATION_NAME",
							Value: "xxx",
						},
						{
							Name:  "DICE_WORKSPACE",
							Value: "xxx",
						},
						{
							Name:  "DICE_RUNTIME_NAME",
							Value: "xxx",
						},
						{
							Name:  "DICE_RUNTIME_ID",
							Value: "xxx",
						},
						{
							Name:  "DICE_APPLICATION_ID",
							Value: "xxx",
						},
						{
							Name:  "DICE_ORG_ID",
							Value: "xxx",
						},
						{
							Name:  "DICE_PROJECT_ID",
							Value: "xxx",
						},
						{
							Name:  "DICE_DEPLOYMENT_ID",
							Value: "xxx",
						},
						{
							Name:  "TERMINUS_LOG_KEY",
							Value: "xxx",
						},
						{
							Name:  "MONITOR_LOG_KEY",
							Value: "xxx",
						},
						{
							Name:  "MSP_ENV_ID",
							Value: "xxx",
						},
						{
							Name:  "MSP_LOG_ATTACH",
							Value: "xxx",
						},
						{
							Name:  "MONITOR_LOG_COLLECTOR",
							Value: "xxx",
						},
						{
							Name:  "TERMINUS_KEY",
							Value: "xxx",
						},
					},
				},
				podAnnotations: make(map[string]string),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestDereferenceEnvs(t *testing.T) {
	var d = new(appsv1.Deployment)
	d.Spec.Template.Spec.Containers = []apiv1.Container{
		{Env: []apiv1.EnvVar{
			{Name: "ENV_A", Value: "homework"},
			{Name: "ENV_B", Value: "do ${env.ENV_A}"},
		}},
	}
	if err := DereferenceEnvs(d); err != nil {
		t.Fatal(err)
	}
	for _, env := range d.Spec.Template.Spec.Containers[0].Env {
		t.Logf("Name: %s, Value: %s", env.Name, env.Value)
	}
}
