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
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/constraints"
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

	// 将原始服务名与statefulset下的实例名关联起来
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

	// 取其中一个服务
	service := &info.sg.Services[0]

	// 1核等于1000m
	cpu := fmt.Sprintf("%.fm", service.Resources.Cpu*1000)
	// 1Mi=1024K=1024x1024字节
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
	// 当前我们的Pod中只设定一个业务容器
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
	// 强制拉镜像
	//container.ImagePullPolicy = apiv1.PullAlways

	// 设置 volume
	if err := k.setVolume(set, container, service); err != nil {
		return err
	}

	// 配置健康检查
	SetHealthCheck(container, service)

	if len(service.Cmd) > 0 {
		container.Command = []string{"sh", "-c", service.Cmd}
	}

	//根据环境设置超分比
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

	// 根据超卖比，设置细粒度的CPU
	if err := k.SetFineGrainedCPU(container, info.sg.Extra, cpuSubscribeRatio); err != nil {
		return err
	}
	if err := k.SetOverCommitMem(container, memSubscribeRatio); err != nil {
		return err
	}
	// 设置 statefulset 环境变量
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

		// 名字构成 '[a-z0-9]([-a-z0-9]*[a-z0-9])?'
		name := strutil.Concat("hostpath-", strconv.Itoa(i))

		// 没有以绝对路径开头的 hostPath 是用于老的 volume 接口中申请本地盘资源
		if !strings.HasPrefix(hostPath, "/") {
			//hostPath = strutil.Concat("/mnt/k8s/", hostPath)
			k.requestLocalVolume(set, container, bind)
			continue
		}

		// k8s hostPath 实现挂盘
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
	// Bind 里全部使用 hostPath,  宿主机路径挂载到容器中
	return k.setBind(set, container, service)

	// 新的 volume 接口
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

// 新版 volume 接口
func configNewVolume(set *appsv1.StatefulSet, container *apiv1.Container, service *apistructs.Service) {
	if len(service.Volumes) == 0 {
		return
	}
	// statefulset 的 volume 使用本地盘或者 nas 网盘
	// 本地盘用 hostPath 来模拟，在 statefulset 的实例被调度到不同实例的前提下
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
		// 经过处理后的环境变量的 key 其长度都应该大于3
		if len(k) <= 3 {
			continue
		}
		container.Env = append(container.Env,
			// 已添加前缀 N0_, N1_, N2_, 等环境变量
			apiv1.EnvVar{
				Name:  k,
				Value: v,
			})
	}
	// 加上 K8S 标识
	container.Env = append(container.Env,
		apiv1.EnvVar{
			Name:  "IS_K8S",
			Value: "true",
		})
	// 加上 namespace 标识
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

// statefulset 名字定义为用户设置的 id
func statefulsetName(sg *apistructs.ServiceGroup) string {
	statefulName, ok := getGroupID(&sg.Services[0])
	if !ok {
		logrus.Errorf("failed to get groupID from service labels, name: %s", sg.ID)
		return sg.ID
	}
	return statefulName
}

// ParseJobHostBindTemplate 对 hostPath 进行模版解析，转换成 cluster info 值
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
