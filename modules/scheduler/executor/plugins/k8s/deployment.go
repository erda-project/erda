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
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/constraints"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// DefaultServiceDNSSuffix k8s service dns 固定后缀
	DefaultServiceDNSSuffix = "svc.cluster.local"
	shardDirSuffix          = "-shard-dir"
	sidecarNamePrefix       = "sidecar-"
	EnableServiceLinks      = "ENABLE_SERVICE_LINKS"
)

func (k *Kubernetes) createDeployment(service *apistructs.Service, sg *apistructs.ServiceGroup) error {
	deployment, err := k.newDeployment(service, sg)
	if err != nil {
		return errors.Errorf("failed to generate deployment struct, name: %s, (%v)", service.Name, err)
	}

	return k.deploy.Create(deployment)
}

func (k *Kubernetes) getDeploymentStatus(service *apistructs.Service) (apistructs.StatusDesc, error) {
	var statusDesc apistructs.StatusDesc
	// in version 1.10.3, the following two apis are equal
	// http://localhost:8080/apis/extensions/v1beta1/namespaces/default/deployments/myk8stest6
	// http://localhost:8080/apis/extensions/v1beta1/namespaces/default/deployments/myk8stest6/status
	deploymentName := getDeployName(service)

	deployment, err := k.getDeployment(service.Namespace, deploymentName)
	if err != nil {
		return statusDesc, err
	}
	status := deployment.Status
	// 刚刚开始创建的时候可能获取到该状态
	if len(status.Conditions) == 0 {
		statusDesc.Status = apistructs.StatusUnknown
		statusDesc.LastMessage = "cannot get statusDesc condition"
		return statusDesc, nil
	}

	for _, c := range status.Conditions {
		if c.Type == k8sapi.DeploymentReplicaFailure && c.Status == "True" {
			statusDesc.Status = apistructs.StatusFailing
			return statusDesc, nil
		}
		if c.Type == k8sapi.DeploymentAvailable && c.Status == "False" {
			statusDesc.Status = apistructs.StatusFailing
			return statusDesc, nil
		}
	}

	statusDesc.Status = apistructs.StatusUnknown

	if status.Replicas == status.ReadyReplicas &&
		status.Replicas == status.AvailableReplicas &&
		status.Replicas == status.UpdatedReplicas {
		if status.Replicas > 0 {
			statusDesc.Status = apistructs.StatusReady
			statusDesc.LastMessage = fmt.Sprintf("deployment(%s) is running", deployment.Name)
		} else if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
			statusDesc.Status = apistructs.StatusReady
		} else {
			statusDesc.LastMessage = fmt.Sprintf("deployment(%s) replica is 0, been deleting", deployment.Name)
		}
	}
	return statusDesc, nil
}

func (k *Kubernetes) getDeployment(namespace, name string) (*appsv1.Deployment, error) {
	return k.deploy.Get(namespace, name)
}

func (k *Kubernetes) putDeployment(deployment *appsv1.Deployment) error {
	return k.deploy.Put(deployment)
}

func (k *Kubernetes) deleteDeployment(namespace, name string) error {
	return k.deploy.Delete(namespace, name)
}

// AddContainersEnv 新增容器环境变量
func (k *Kubernetes) AddContainersEnv(containers []apiv1.Container, service *apistructs.Service, sg *apistructs.ServiceGroup) error {
	var envs []apiv1.EnvVar
	ns := MakeNamespace(sg)
	serviceName := service.Name
	if sg.ProjectNamespace != "" {
		ns = sg.ProjectNamespace
		serviceName = service.Env[ProjectNamespaceServiceNameNameKey]
	}

	// 用户注入的环境变量
	if len(service.Env) > 0 {
		for k, v := range service.Env {
			envs = append(envs, apiv1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	addEnv := func(svc *apistructs.Service, envs *[]apiv1.EnvVar, useClusterIP bool) error {
		var err error
		// use SHORT dns, service's name is equal to SHORT dns
		host := strings.Join([]string{serviceName, svc.Namespace, DefaultServiceDNSSuffix}, ".")
		if useClusterIP {
			host, err = k.getClusterIP(svc.Namespace, serviceName)
			if err != nil {
				return err
			}
		}
		// 添加{serviceName}_HOST
		*envs = append(*envs, apiv1.EnvVar{
			Name:  makeEnvVariableName(serviceName) + "_HOST",
			Value: host,
		})

		// {serviceName}_PORT 指代第一个端口
		if len(svc.Ports) > 0 {
			*envs = append(*envs, apiv1.EnvVar{
				Name:  makeEnvVariableName(serviceName) + "_PORT",
				Value: strconv.Itoa(svc.Ports[0].Port),
			})
		}

		// 有多个端口的情况下，依次使用{serviceName}_PORT0,{serviceName}_PORT1,...
		for i, port := range svc.Ports {
			*envs = append(*envs, apiv1.EnvVar{
				Name:  makeEnvVariableName(serviceName) + "_PORT" + strconv.Itoa(i),
				Value: strconv.Itoa(port.Port),
			})
		}
		return nil
	}

	// 所有容器都有所有service的环境变量
	if sg.ServiceDiscoveryMode == "GLOBAL" {
		// 由于某些服务可能尚未创建，所以注入dns
		// TODO: 也可以先将runtime的所有service的k8s service先创建出来
		for _, svc := range sg.Services {
			// 无端口暴露的服务无{serviceName}_HOST,{serviceName}_PORT等环境变量
			if len(svc.Ports) == 0 {
				continue
			}
			svc.Namespace = ns
			// dns环境变量
			if err := addEnv(&svc, &envs, false); err != nil {
				return err
			}
		}
	} else {
		// 按约定注入依赖的服务的环境变量，比如服务A依赖服务B，
		// 则将{B}_HOST, {B}_PORT 注入到对A可见的容器环境变量中
		var backendURLEnv apiv1.EnvVar
		for _, name := range service.Depends {
			var dependSvc *apistructs.Service
			for _, svc := range sg.Services {
				if svc.Name != name {
					continue
				}
				dependSvc = &svc
				break
			}
			// this would not happen as if this situations exist,
			// we would find out it in the stage of util.ParseServiceDependency
			if dependSvc == nil {
				return errors.Errorf("failed to find service in runtime, servicename: %s", name)
			}

			if len(dependSvc.Ports) == 0 {
				continue
			}

			// 注入{B}_HOST, {B}_PORT等
			if err := addEnv(dependSvc, &envs, false); err != nil {
				return err
			}

			// 适配上层逻辑, 如果IS_ENDPOINT被赋为true, 增加BACKEND_URL环境变量
			if service.Labels["IS_ENDPOINT"] == "true" && len(dependSvc.Ports) > 0 {
				backendIP := strings.Join([]string{getServiceName(dependSvc), dependSvc.Namespace, DefaultServiceDNSSuffix}, ".")
				backendPort := dependSvc.Ports[0].Port
				backendURLEnv = apiv1.EnvVar{
					Name:  "BACKEND_URL",
					Value: backendIP + ":" + strconv.Itoa(backendPort),
				}
			}
		}

		if len(backendURLEnv.Name) > 0 {
			envs = append(envs, backendURLEnv)
			logrus.Debugf("got service BACKEND_URL, service: %s, namespace: %s, url: %s",
				serviceName, service.Namespace, backendURLEnv.Value)
		}

		// 注入服务自身的环境变量，即{A}_HOST, {A}_PORT等
		if len(service.Ports) > 0 {
			if err := addEnv(service, &envs, false); err != nil {
				return err
			}
		}
	}

	// 加上 K8S 标识
	envs = append(envs, apiv1.EnvVar{
		Name:  "IS_K8S",
		Value: "true",
	})

	// 加上 namespace 标识
	envs = append(envs, apiv1.EnvVar{
		Name:  "DICE_NAMESPACE",
		Value: ns,
	})

	if service.NewHealthCheck != nil && service.NewHealthCheck.HttpHealthCheck != nil {
		envs = append(envs, apiv1.EnvVar{
			Name:  "DICE_HTTP_HEALTHCHECK_PATH",
			Value: service.NewHealthCheck.HttpHealthCheck.Path,
		})
	}
	// 加上 POD_IP 及 HOST_IP
	envs = append(envs, apiv1.EnvVar{
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
	})

	// Add POD_NAME
	envs = append(envs, apiv1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})

	// Add POD_UUID
	envs = append(envs, apiv1.EnvVar{
		Name: "POD_UUID",
		ValueFrom: &apiv1.EnvVarSource{
			FieldRef: &apiv1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.uid",
			},
		},
	})

	// 加上 SELF_URL、SELF_HOST、SELF_PORT
	selfHost := strings.Join([]string{serviceName, service.Namespace, DefaultServiceDNSSuffix}, ".")
	envs = append(envs, apiv1.EnvVar{
		Name:  "SELF_HOST",
		Value: selfHost,
	})
	for i, port := range service.Ports {
		portStr := strconv.Itoa(port.Port)
		if i == 0 {
			// TODO: we should deprecate this SELF_URL
			// TODO: we don't care what http is
			// special port env, SELF_PORT == SELF_PORT0
			envs = append(envs, apiv1.EnvVar{
				Name:  "SELF_PORT",
				Value: portStr,
			}, apiv1.EnvVar{
				Name:  "SELF_URL",
				Value: "http://" + selfHost + ":" + portStr,
			})

		}
		envs = append(envs, apiv1.EnvVar{
			Name:  "SELF_PORT" + strconv.Itoa(i),
			Value: portStr,
		})
	}

	for i, container := range containers {
		requestmem, _ := container.Resources.Requests.Memory().AsInt64()
		limitmem, _ := container.Resources.Limits.Memory().AsInt64()
		envs = append(envs,
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

		containers[i].Env = envs
	}
	return nil
}

func podAnnotations(service *apistructs.Service, podannotations map[string]string) {
	// 默认关闭 inject，不受 ns 开启 inject 的影响
	podannotations["sidecar.istio.io/inject"] = "false"
	// 开启 mesh，并且启动了 security traffic 的情况下，需要劫持HTTP的健康检查
	if service.MeshEnable != nil {
		if *service.MeshEnable {
			podannotations["sidecar.istio.io/inject"] = "true"
			if service.TrafficSecurity.Mode == "https" && service.NewHealthCheck != nil &&
				service.NewHealthCheck.HttpHealthCheck != nil {
				podannotations["sidecar.istio.io/rewriteAppHTTPProbers"] = "true"
			}
		}
	}
}

func (k *Kubernetes) newDeployment(service *apistructs.Service, sg *apistructs.ServiceGroup) (*appsv1.Deployment, error) {
	deploymentName := getDeployName(service)
	enableServiceLinks := false
	if _, ok := sg.Labels[EnableServiceLinks]; ok {
		enableServiceLinks = true
	}
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: service.Namespace,
			Labels:    make(map[string]string),
		},
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: func(i int32) *int32 { return &i }(int32(3)),
			Replicas:             func(i int32) *int32 { return &i }(int32(service.Scale)),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   deploymentName,
					Labels: make(map[string]string),
				},
				Spec: apiv1.PodSpec{
					EnableServiceLinks:    func(enable bool) *bool { return &enable }(enableServiceLinks),
					ShareProcessNamespace: func(b bool) *bool { return &b }(false),
					Tolerations:           toleration.GenTolerations(),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": service.Name},
			},
		},
	}

	if v := k.options["FORCE_BLUE_GREEN_DEPLOY"]; v != "true" &&
		(strutil.ToUpper(service.Env["DICE_WORKSPACE"]) == apistructs.DevWorkspace.String() ||
			strutil.ToUpper(service.Env["DICE_WORKSPACE"]) == apistructs.TestWorkspace.String()) {
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{Type: "Recreate"}
	}

	affinity := constraintbuilders.K8S(&sg.ScheduleInfo2, service, []constraints.PodLabelsForAffinity{
		{PodLabels: map[string]string{"app": service.Name}}}, k).Affinity
	deployment.Spec.Template.Spec.Affinity = &affinity

	// 注入 hosts
	deployment.Spec.Template.Spec.HostAliases = ConvertToHostAlias(service.Hosts)

	// 1核等于1000m
	cpu := fmt.Sprintf("%.fm", service.Resources.Cpu*1000)
	// 1Mi=1024K=1024x1024字节
	memory := fmt.Sprintf("%.fMi", service.Resources.Mem)

	container := apiv1.Container{
		// TODO, container name e.g. redis-1528180634
		Name:  service.Name,
		Image: service.Image,
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse(cpu),
				apiv1.ResourceMemory: resource.MustParse(memory),
			},
		},
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
	if err := k.SetFineGrainedCPU(&container, sg.Extra, cpuSubscribeRatio); err != nil {
		return nil, err
	}

	if err := k.SetOverCommitMem(&container, memSubscribeRatio); err != nil {
		return nil, err
	}

	// 生成 sidecars 容器配置
	sidecars := k.generateSidecarContainers(service.SideCars)

	// 生成 initcontainer 配置
	initcontainers := k.generateInitContainer(service.InitContainer)

	containers := []apiv1.Container{container}
	containers = append(containers, sidecars...)
	deployment.Spec.Template.Spec.Containers = containers
	if len(initcontainers) > 0 {
		deployment.Spec.Template.Spec.InitContainers = initcontainers
	}
	//for k, v := range service.Labels {
	// TODO: temporary modifications
	//if k != "HAPROXY_0_VHOST" {
	//	deployment.Metadata.Labels[k] = v
	//	deployment.Spec.Template.Metadata.Labels[k] = v
	//}
	//}

	// k8s deployment 必须得有app的label来辖管pod
	deployment.Labels["app"] = service.Name
	deployment.Spec.Template.Labels["app"] = service.Name

	setDeploymentLabels(service, deployment)

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	podAnnotations(service, deployment.Spec.Template.Annotations)

	// 按当前的设定，一个pod中只有一个用户的容器
	if service.Cmd != "" {
		for i := range containers {
			//TODO:
			//cmds := strings.Split(service.Cmd, " ")
			cmds := []string{"sh", "-c", service.Cmd}
			containers[i].Command = cmds
		}
	}

	// 配置健康检查
	SetHealthCheck(&deployment.Spec.Template.Spec.Containers[0], service)

	if err := k.AddContainersEnv(containers, service, sg); err != nil {
		return nil, err
	}

	// TODO: 删除这段逻辑
	// 美孚临时需求:
	// 将 "secret" namespace 下的 secret 注入业务容器

	secrets, err := k.CopyDiceSecrets("secret", service.Namespace)
	if err != nil {
		logrus.Errorf("failed to copy secret: %v", err)
		return nil, err
	}
	secretvolumes := []apiv1.Volume{}
	secretvolmounts := []apiv1.VolumeMount{}
	for _, secret := range secrets {
		secretvol, volmount := k.SecretVolume(&secret)
		secretvolumes = append(secretvolumes, secretvol)
		secretvolmounts = append(secretvolmounts, volmount)
	}

	k.AddPodMountVolume(service, &deployment.Spec.Template.Spec, secretvolmounts, secretvolumes)
	k.AddSpotEmptyDir(&deployment.Spec.Template.Spec)

	logrus.Debugf("show k8s deployment, name: %s, deployment: %+v", deploymentName, deployment)
	return deployment, nil
}

func (k *Kubernetes) generateInitContainer(initcontainers map[string]diceyml.InitContainer) []apiv1.Container {
	containers := []apiv1.Container{}
	if initcontainers == nil {
		return containers
	}
	for name, initcontainer := range initcontainers {
		reqCPU := fmt.Sprintf("%.fm", k.CPUOvercommit(initcontainer.Resources.CPU)*1000)
		limitCPU := fmt.Sprintf("%.fm", initcontainer.Resources.CPU*1000)
		memory := fmt.Sprintf("%.dMi", initcontainer.Resources.Mem)

		sc := apiv1.Container{
			Name:  name,
			Image: initcontainer.Image,
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(reqCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(limitCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
			},
			Command: []string{"sh", "-c", initcontainer.Cmd},
		}
		for i, dir := range initcontainer.SharedDirs {
			emptyDirVolumeName := fmt.Sprintf("%s-%d", name, i)
			dstMount := apiv1.VolumeMount{
				Name:      emptyDirVolumeName,
				MountPath: dir.SideCar,
				ReadOnly:  false, // rw
			}
			sc.VolumeMounts = append(sc.VolumeMounts, dstMount)
		}
		containers = append(containers, sc)
	}

	return containers
}

func (k *Kubernetes) generateSidecarContainers(sidecars map[string]*diceyml.SideCar) []apiv1.Container {
	var containers []apiv1.Container

	if len(sidecars) == 0 {
		return nil
	}

	for name, sidecar := range sidecars {
		// 1核等于1000m
		reqCPU := fmt.Sprintf("%.fm", k.CPUOvercommit(sidecar.Resources.CPU)*1000)
		limitCPU := fmt.Sprintf("%.fm", sidecar.Resources.CPU*1000)
		// 1Mi=1024K=1024x1024字节
		memory := fmt.Sprintf("%.dMi", sidecar.Resources.Mem)

		sc := apiv1.Container{
			Name:  strutil.Concat(sidecarNamePrefix, name),
			Image: sidecar.Image,
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(reqCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse(limitCPU),
					apiv1.ResourceMemory: resource.MustParse(memory),
				},
			},
		}

		// 不塞入平台环境变量(DICE_*)，防止实例列表被采集
		for k, v := range sidecar.Envs {
			sc.Env = append(sc.Env, apiv1.EnvVar{
				Name:  k,
				Value: v,
			})
		}

		// sidecar 与业务容器共享目录
		for _, dir := range sidecar.SharedDirs {
			emptyDirVolumeName := strutil.Concat(name, shardDirSuffix)
			dstMount := apiv1.VolumeMount{
				Name:      emptyDirVolumeName,
				MountPath: dir.SideCar,
				ReadOnly:  false, // rw
			}
			sc.VolumeMounts = append(sc.VolumeMounts, dstMount)
		}
		containers = append(containers, sc)
	}

	return containers
}

func makeEnvVariableName(str string) string {
	return strings.ToUpper(strings.Replace(str, "-", "_", -1))
}

// AddPodMountVolume 新增 pod volume 配置
func (k *Kubernetes) AddPodMountVolume(service *apistructs.Service, podSpec *apiv1.PodSpec,
	secretvolmounts []apiv1.VolumeMount, secretvolumes []apiv1.Volume) error {

	podSpec.Volumes = make([]apiv1.Volume, 0)

	// 注意上面提到的设定，一个pod中只有一个container
	podSpec.Containers[0].VolumeMounts = make([]apiv1.VolumeMount, 0)

	// get cluster info
	clusterInfo, err := k.ClusterInfo.Get()
	if err != nil {
		return errors.Errorf("failed to get cluster info, clusterName: %s, (%v)", k.clusterName, err)
	}

	// hostPath类型
	for i, bind := range service.Binds {
		if bind.HostPath == "" || bind.ContainerPath == "" {
			continue
		}
		// 名字构成 '[a-z0-9]([-a-z0-9]*[a-z0-9])?'
		name := "volume" + "-bind-" + strconv.Itoa(i)

		hostPath, err := ParseJobHostBindTemplate(bind.HostPath, clusterInfo)
		if err != nil {
			return err
		}
		if !strutil.HasPrefixes(hostPath, "/") {
			pvcName := strings.Replace(hostPath, "_", "-", -1)
			sc := "dice-local-volume"
			if err := k.pvc.CreateIfNotExists(&apiv1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s", service.Name, pvcName),
					Namespace: service.Namespace,
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
			}); err != nil {
				return err
			}

			podSpec.Volumes = append(podSpec.Volumes,
				apiv1.Volume{
					Name: name,
					VolumeSource: apiv1.VolumeSource{
						PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
							ClaimName: fmt.Sprintf("%s-%s", service.Name, pvcName),
						},
					},
				})

			podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts,
				apiv1.VolumeMount{
					Name:      name,
					MountPath: bind.ContainerPath,
					ReadOnly:  bind.ReadOnly,
				})
			continue
		}

		podSpec.Volumes = append(podSpec.Volumes,
			apiv1.Volume{
				Name: name,
				VolumeSource: apiv1.VolumeSource{
					HostPath: &apiv1.HostPathVolumeSource{
						Path: hostPath,
					},
				},
			})

		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts,
			apiv1.VolumeMount{
				Name:      name,
				MountPath: bind.ContainerPath,
				ReadOnly:  bind.ReadOnly,
			})
	}

	// 配置业务容器 sidecar 共享目录
	for name, sidecar := range service.SideCars {
		for _, dir := range sidecar.SharedDirs {
			emptyDirVolumeName := strutil.Concat(name, shardDirSuffix)

			srcMount := apiv1.VolumeMount{
				Name:      emptyDirVolumeName,
				MountPath: dir.Main,
				ReadOnly:  false, // rw
			}
			// 业务主容器
			podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, srcMount)

			podSpec.Volumes = append(podSpec.Volumes, apiv1.Volume{
				Name: emptyDirVolumeName,
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			})
		}
	}
	if service.InitContainer != nil {
		for name, initc := range service.InitContainer {
			for i, dir := range initc.SharedDirs {
				name := fmt.Sprintf("%s-%d", name, i)
				srcMount := apiv1.VolumeMount{
					Name:      name,
					MountPath: dir.Main,
					ReadOnly:  false,
				}
				podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, srcMount)
				podSpec.Volumes = append(podSpec.Volumes, apiv1.Volume{
					Name: name,
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{},
					},
				})
			}
		}
	}

	podSpec.Volumes = append(podSpec.Volumes, secretvolumes...)
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, secretvolmounts...)

	return nil
}
func (k *Kubernetes) AddSpotEmptyDir(podSpec *apiv1.PodSpec) {
	podSpec.Volumes = append(podSpec.Volumes, apiv1.Volume{
		Name:         "spot-emptydir",
		VolumeSource: apiv1.VolumeSource{EmptyDir: &apiv1.EmptyDirVolumeSource{}},
	})

	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, apiv1.VolumeMount{
		Name:      "spot-emptydir",
		MountPath: "/tmp",
	})
}

func getDeployName(service *apistructs.Service) string {
	if service.Env[ProjectNamespace] == "true" {
		return service.Env[ProjectNamespaceServiceNameNameKey]
	}
	return service.Name
}

func setDeploymentLabels(service *apistructs.Service, deployment *appsv1.Deployment) {
	if v, ok := service.Env[ProjectNamespace]; ok && v == "true" {
		deployment.Spec.Selector.MatchLabels[LabelServiceGroupID] = service.Env[KeyServiceGroupID]
		deployment.Spec.Template.Labels[LabelServiceGroupID] = service.Env[KeyServiceGroupID]
		deployment.Labels[LabelServiceGroupID] = service.Env[KeyServiceGroupID]
	}
}

func ConvertToHostAlias(hosts []string) []apiv1.HostAlias {
	var r []apiv1.HostAlias
	for _, host := range hosts {
		splitRes := strings.Fields(host)
		if len(splitRes) < 2 {
			continue
		}
		r = append(r, apiv1.HostAlias{
			IP:        splitRes[0],
			Hostnames: splitRes[1:],
		})
	}
	return r
}
