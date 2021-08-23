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
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders/constraints"
	"github.com/erda-project/erda/pkg/strutil"
)

func (k *Kubernetes) createStatefulSet(info StatefulsetInfo) error {
	statefulName := statefulsetName(info.sg)

	set := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        statefulName,
			Namespace:   info.namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}

	// Associate the original service name with the instance name under statefulset
	for k, v := range info.annotations {
		set.Annotations[k] = v
	}

	set.Spec = appsv1.StatefulSetSpec{
		RevisionHistoryLimit: func(i int32) *int32 { return &i }(int32(3)),
		Replicas:             func(i int32) *int32 { return &i }(int32(len(info.sg.Services))),
		ServiceName:          statefulName,
	}

	set.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"app": statefulName},
	}

	// Take one of the services
	service := &info.sg.Services[0]

	cpu := fmt.Sprintf("%.fm", service.Resources.Cpu*1000)
	memory := fmt.Sprintf("%.fMi", service.Resources.Mem)

	affinity := constraintbuilders.K8S(&info.sg.ScheduleInfo2, service, []constraints.PodLabelsForAffinity{
		{
			PodLabels: map[string]string{"app": statefulName},
		}}, k).Affinity

	set.Spec.Template = apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: info.namespace,
			Labels:    map[string]string{"app": statefulName},
		},
		Spec: apiv1.PodSpec{
			EnableServiceLinks:    func(enable bool) *bool { return &enable }(false),
			ShareProcessNamespace: func(b bool) *bool { return &b }(false),
			Tolerations:           toleration.GenTolerations(),
		},
	}
	set.Spec.Template.Spec.Affinity = &affinity
	// Currently only one business container is set in our Pod
	container := &apiv1.Container{
		Name:  statefulName,
		Image: service.Image,
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse(cpu),
				apiv1.ResourceMemory: resource.MustParse(memory),
			},
		},
	}
	// Forced to pull the image
	//container.ImagePullPolicy = apiv1.PullAlways

	// setting volume
	if err := k.setVolume(set, container, service); err != nil {
		return err
	}

	// configure health check
	SetHealthCheck(container, service)

	if len(service.Cmd) > 0 {
		container.Command = []string{"sh", "-c", service.Cmd}
	}

	// Set the over-score ratio according to the environment
	cpuSubscribeRatio := k.cpuSubscribeRatio
	memSubscribeRatio := k.memSubscribeRatio
	switch strutil.ToUpper(service.Env["DICE_WORKSPACE"]) {
	case "DEV":
		cpuSubscribeRatio = k.devCpuSubscribeRatio
		memSubscribeRatio = k.devMemSubscribeRatio
	case "TEST":
		cpuSubscribeRatio = k.testCpuSubscribeRatio
		memSubscribeRatio = k.testMemSubscribeRatio
	case "STAGING":
		cpuSubscribeRatio = k.stagingCpuSubscribeRatio
		memSubscribeRatio = k.stagingMemSubscribeRatio
	}

	// Set fine-grained CPU based on the oversold ratio
	if err := k.SetFineGrainedCPU(container, info.sg.Extra, cpuSubscribeRatio); err != nil {
		return err
	}
	if err := k.SetOverCommitMem(container, memSubscribeRatio); err != nil {
		return err
	}
	// Set the statefulset environment variable
	setEnv(container, info.envs, info.sg, info.namespace)

	set.Spec.Template.Spec.Containers = []apiv1.Container{*container}
	return k.sts.Create(set)
}

func (k *Kubernetes) setBind(set *appsv1.StatefulSet, container *apiv1.Container, service *apistructs.Service) error {
	for i, bind := range service.Binds {
		if bind.HostPath == "" || bind.ContainerPath == "" {
			continue
		}

		clusterInfo, err := k.ClusterInfo.Get()
		if err != nil {
			return err
		}
		hostPath, err := ParseJobHostBindTemplate(bind.HostPath, clusterInfo)
		if err != nil {
			return err
		}

		// Name formation: '[a-z0-9]([-a-z0-9]*[a-z0-9])?'
		name := strutil.Concat("hostpath-", strconv.Itoa(i))

		// The hostPath that does not start with an absolute path is used to apply for local disk resources in the old volume interface
		if !strings.HasPrefix(hostPath, "/") {
			//hostPath = strutil.Concat("/mnt/k8s/", hostPath)
			k.requestLocalVolume(set, container, bind)
			continue
		}

		// k8s hostPath Achieve listing
		set.Spec.Template.Spec.Volumes = append(set.Spec.Template.Spec.Volumes, apiv1.Volume{
			Name: name,
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: hostPath,
				},
			},
		})

		container.VolumeMounts = append(container.VolumeMounts,
			apiv1.VolumeMount{
				Name:      name,
				MountPath: bind.ContainerPath,
				ReadOnly:  bind.ReadOnly,
			})
	}

	return nil
}

func (k *Kubernetes) setVolume(set *appsv1.StatefulSet, container *apiv1.Container, service *apistructs.Service) error {
	// HostPath is all used in Bind, and the host path is mounted to the container
	return k.setBind(set, container, service)

	// new volume interface
	// configNewVolume(set, container, service)
}

func (k *Kubernetes) requestLocalVolume(set *appsv1.StatefulSet, container *apiv1.Container, bind apistructs.ServiceBind) {
	logrus.Infof("in requestLocalVolume, statefulset name: %s, hostPath: %s", set.Name, bind.HostPath)

	sc := localStorage

	hostPath := bind.HostPath
	pvcName := strings.Replace(hostPath, "_", "-", -1)
	set.Spec.VolumeClaimTemplates = append(set.Spec.VolumeClaimTemplates, apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvcName,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
			StorageClassName: &sc,
		},
	})

	container.VolumeMounts = append(container.VolumeMounts,
		apiv1.VolumeMount{
			Name:      pvcName,
			MountPath: bind.ContainerPath,
			ReadOnly:  bind.ReadOnly,
		})
}

// new volume interface
func configNewVolume(set *appsv1.StatefulSet, container *apiv1.Container, service *apistructs.Service) {
	if len(service.Volumes) == 0 {
		return
	}
	// The volume of statefulset uses local disk or nas network disk
	// The local disk is simulated by hostPath, under the premise that the instances of statefulset are scheduled to different instances
	for _, vol := range service.Volumes {
		nas := 0
		local := 0
		var (
			name string
			path string
		)
		switch vol.VolumeType {
		case apistructs.LocalVolume:
			name = strutil.Concat("localvolume-", strconv.Itoa(local))
			path = strutil.Concat("/mnt/dice/volumes/", vol.VolumePath)
			local++
		case apistructs.NasVolume:
			name = strutil.Concat("nas-", strconv.Itoa(nas))
			path = vol.VolumePath
			nas++
		}

		container.VolumeMounts = []apiv1.VolumeMount{
			{
				Name:      name,
				MountPath: vol.ContainerPath,
			},
		}

		set.Spec.Template.Spec.Volumes = append(set.Spec.Template.Spec.Volumes,
			apiv1.Volume{
				Name: name,
				VolumeSource: apiv1.VolumeSource{
					HostPath: &apiv1.HostPathVolumeSource{
						Path: path,
					},
				},
			},
		)
	}
}

func setEnv(container *apiv1.Container, allEnv map[string]string, sg *apistructs.ServiceGroup, ns string) {
	// copy all env variable
	for k, v := range allEnv {
		// The key length of the processed environment variable should be greater than 3
		if len(k) <= 3 {
			continue
		}
		container.Env = append(container.Env,
			// The prefixes N0_, N1_, N2_, and other environment variables have been added
			apiv1.EnvVar{
				Name:  k,
				Value: v,
			})
	}
	// add K8S label
	container.Env = append(container.Env,
		apiv1.EnvVar{
			Name:  "IS_K8S",
			Value: "true",
		})
	// add namespace label
	container.Env = append(container.Env,
		apiv1.EnvVar{
			Name:  "DICE_NAMESPACE",
			Value: ns,
		})
	if len(sg.Services) > 0 {
		service := sg.Services[0]
		if len(service.Ports) >= 1 {
			container.Env = append(container.Env,
				apiv1.EnvVar{Name: "SELF_PORT", Value: fmt.Sprintf("%d", service.Ports[0].Port)})
		}
		for i, port := range service.Ports {
			container.Env = append(container.Env,
				apiv1.EnvVar{
					Name:  fmt.Sprintf("SELF_PORT%d", i),
					Value: fmt.Sprintf("%d", port.Port),
				})
		}

		requestmem, _ := container.Resources.Requests.Memory().AsInt64()
		limitmem, _ := container.Resources.Limits.Memory().AsInt64()
		container.Env = append(container.Env,
			apiv1.EnvVar{
				Name:  "DICE_CPU_ORIGIN",
				Value: fmt.Sprintf("%f", service.Resources.Cpu)},
			apiv1.EnvVar{
				Name:  "DICE_MEM_ORIGIN",
				Value: fmt.Sprintf("%f", service.Resources.Mem)},
			apiv1.EnvVar{
				Name:  "DICE_CPU_REQUEST",
				Value: container.Resources.Requests.Cpu().AsDec().String()},
			apiv1.EnvVar{
				Name:  "DICE_MEM_REQUEST",
				Value: fmt.Sprintf("%d", requestmem/1024/1024)},
			apiv1.EnvVar{
				Name:  "DICE_CPU_LIMIT",
				Value: container.Resources.Limits.Cpu().AsDec().String()},
			apiv1.EnvVar{
				Name:  "DICE_MEM_LIMIT",
				Value: fmt.Sprintf("%d", limitmem/1024/1024)},
		)
	}

	container.Env = append(container.Env, apiv1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}, apiv1.EnvVar{
		Name: "HOST_IP",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	}, apiv1.EnvVar{
		Name: "SELF_HOST",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}, apiv1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	}, apiv1.EnvVar{
		Name: "POD_UUID",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.uid",
			},
		},
	})
}

func convertStatus(status apiv1.PodStatus) apistructs.StatusCode {
	switch status.Phase {
	case apiv1.PodRunning:
		for _, cond := range status.Conditions {
			if cond.Status == "False" {
				return apistructs.StatusProgressing
			}
		}
		return apistructs.StatusReady
	}
	return apistructs.StatusProgressing
}

// statefulset The name is defined as set by the user id
func statefulsetName(sg *apistructs.ServiceGroup) string {
	statefulName, ok := getGroupID(&sg.Services[0])
	if !ok {
		logrus.Errorf("failed to get groupID from service labels, name: %s", sg.ID)
		return sg.ID
	}
	return statefulName
}

// ParseJobHostBindTemplate Analyze the hostPath template and convert it to the cluster info value
func ParseJobHostBindTemplate(hostPath string, clusterInfo map[string]string) (string, error) {
	var b bytes.Buffer

	if hostPath == "" {
		return "", errors.New("hostPath is empty")
	}

	t, err := template.New("jobBind").
		Option("missingkey=error").
		Parse(hostPath)
	if err != nil {
		return "", errors.Errorf("failed to parse bind, hostPath: %s, (%v)", hostPath, err)
	}

	err = t.Execute(&b, &clusterInfo)
	if err != nil {
		return "", errors.Errorf("failed to execute bind, hostPath: %s, (%v)", hostPath, err)
	}

	return b.String(), nil
}
